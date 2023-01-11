package partida

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"log"
	"time"

	"gopkg.in/guregu/null.v3"
)

// data format response for an action function return
type RequestResponse struct {
	Code          int         // return operation code
	CustomMessage string      // custom action return message
	Data          interface{} // operation generated data
}

// relevant request data assembled by HTTP handler
type Request struct {

	// general request information
	ID          string      // request identifier for debug
	Logger      *log.Logger // an logger instance
	ContentType string      // in which format the response should be returned json/xml

	// protocol level
	IP      string            // end-user IP address
	Query   map[string]string // GET method query parameters
	Headers map[string]string // request HTTP headers

	path   string // request path with the action name
	method string // HTTP request method
	input  []byte // request input data

	// handler level
	ActionName  null.String            // action name
	Parameters  map[string]interface{} // parsed informed data
	HeaderToken string                 // !!!

	result RequestResponse // action handler result
	action Action          // complete action data

}

// data to be returned to the user by the HTTP handler
type Response struct {
	HTTPCode int               // HTTP response code of the result
	Content  []byte            // response content
	Headers  map[string]string // response headers
}

// structure of the metadata used on API responses
type ResponseMeta struct {
	ID            string            `json:"id"`         // request identifier for debug
	Time          time.Time         `json:"time"`       // request answered when
	Code          int               `json:"code"`       // API response code
	Action        null.String       `json:"action"`     // responsible action
	Message       map[string]string `json:"message"`    // response code messages
	CustomMessage map[string]string `json:"op_message"` // the message returned by the action
}

// call all the validators and action function for the requested action
func (c *Controller) handleRequest(r *Request) Response {
	var err error

	r.Logger = log.New(c.writer, fmt.Sprintf("REQ [%v] ", r.ID), log.LstdFlags|log.Lmsgprefix)
	r.Logger.Printf("handler initiated [path: %v] [method: %v] [ip: %v]", r.path, r.method, r.IP)

	// perform the general request verifications
	err = r.validateRequest(c.actions)
	if err != nil {
		return c.makeResponse(r)
	}

	// perform the user authentication and authorization
	if r.action.Authentication {

		r.Logger.Println("this action require authentication")

		// fetch the token passed on the Authorization header
		err = r.fetchUserToken()
		if err != nil {
			return c.makeResponse(r)
		}

		// fetch the token data based on the passed header
		err = c.backend.PerformUserAuthorization(r)
		if err != nil {
			// !! handle error
			return c.makeResponse(r)
		}

	}

	// check the input data for body params (query/post)
	if len(r.action.Parameters) > 0 {

		r.Logger.Println("this action require body parameters")

		// fetch the token data based on the passed header
		err = r.validateBodyParameters(c.validators)
		if err != nil {
			return c.makeResponse(r)
		}

	}

	// execute the action method function
	err = r.callMethod(c.handlers)
	if err != nil {
		return c.makeResponse(r)
	}

	// call the backend post execution function
	err = c.backend.PerformPostExecutionOperations(r)
	if err != nil {
		return c.makeResponse(r)
	}

	r.Logger.Printf("finished the request handler job")

	return c.makeResponse(r)
}

// asd
func (c *Controller) getCustomMessage(name string) map[string]string {

	var msg map[string]string

	if v, ok := c.messages[name]; ok {
		msg = v
	}

	return msg
}

// update the result on the request and "raise" an error if necessarily
func (r *Request) updateResult(code int, msg string, data interface{}, err error) error {

	// update the request result
	r.result = RequestResponse{Code: code, CustomMessage: msg, Data: data}

	// in case an error has been passed, do a "raise"
	if err != nil {
		return err
	}

	// in case the new API code is not 0 (unsuccessful), do a "raise"
	if code != 0 {
		return ErrRequestHandlerUnsuccessful
	}

	return nil
}

// return an HTTP response for the current request result
func (c *Controller) makeResponse(r *Request) Response {
	var err error

	r.Logger.Printf("starting the response assemble [code: %v] [customMessage: %v]", r.result.Code, r.result.CustomMessage)

	// infer the desired Accept response type if no content-type is already set
	if val, ok := r.Headers["Accept"]; ok {
		if val == "application/xml" && r.ContentType == "" {
			r.Logger.Println("application/xml content type found on \"Accept\" header")
			r.ContentType = "xml"
		}
	}

	// check if the response code exists and fetch its data
	codeData := c.codes[1]

	if v, ok := c.codes[r.result.Code]; ok {
		codeData = v
	}

	// get the custom message
	customMsg := c.getCustomMessage(r.result.CustomMessage)

	// set the CORS, CACHE and content type headers
	var headers map[string]string = map[string]string{
		"Content-Type":                 "application/json; charset=utf-8",
		"Cache-Control":                "max-age=0,private,must-revalidate,no-cache",
		"Access-Control-Allow-Origin":  "*",
		"Access-Control-Allow-Methods": "*",
		"Access-Control-Allow-Headers": "*",
		"Access-Control-Max-Age":       "86400",
	}

	if r.ContentType == "xml" {
		headers["Content-Type"] = "application/xml"
		r.Logger.Println("this response will be return as XML")
	}

	// assemble the request response with the code and provided data
	response := struct {
		XMLName xml.Name     `json:"-"`
		Meta    ResponseMeta `json:"meta" xml:"meta"`
		Data    interface{}  `json:"data" xml:"data"`
	}{
		XMLName: xml.Name{Local: "response"},
		Meta: ResponseMeta{
			ID:            r.ID,
			Time:          time.Now(),
			Code:          r.result.Code,
			Action:        r.ActionName,
			Message:       codeData.Message,
			CustomMessage: customMsg,
		},
		Data: r.result.Data,
	}

	// perform the XML/JSON marshaling of the response
	var content []byte

	if r.ContentType == "xml" {
		content, err = xml.MarshalIndent(response, "", "   ")
	} else {
		content, err = json.Marshal(response)
	}

	if err != nil {
		r.Logger.Fatalf("error while marshaling JSON/XML with response [err: %v]", err)
		return Response{HTTPCode: c.codes[1].HTTPCode, Content: []byte{}, Headers: nil}
	}

	r.Logger.Println("API format response assembled for HTTP return")

	return Response{HTTPCode: codeData.HTTPCode, Content: content, Headers: headers}
}
