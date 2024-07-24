package partida

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"log"
	"time"
)

// HandlerResponse for the general method response structure
type HandlerResponse struct {
	Code string      // return operation code
	Data interface{} // operation generated data
}

// Code for the application default response messages and HTTP codes
type Code struct {
	HTTPCode int               `json:"http"`    // HTTP return code
	Message  map[string]string `json:"message"` // messages from the code
}

// Response for the data to be returned to the user by the HTTP handler
type Response struct {
	HTTPCode int               // HTTP response code of the result
	Content  []byte            // response content
	Headers  map[string]string // response headers
}

// ResponseMeta is the structure of the metadata used on API responses
type ResponseMeta struct {
	Code    string            `json:"code"`    // API response code
	ID      string            `json:"id"`      // request identifier for debug
	Time    time.Time         `json:"time"`    // request answered datetime
	Message map[string]string `json:"message"` // response code messages
}

// Request is the structure for the request metadata, payload and useful contents
type Request struct {
	ID          string      // request identifier
	ContentType string      // format of the response (json or xml) default = json
	Logger      *log.Logger // general request logging

	IP      string            // request initiator IP address
	Query   map[string]string // GET method query parameters
	Headers map[string]string // request HTTP headers
	Path    string            // requested path
	Method  string            // HTTP request verb
	Input   []byte            // input data

	ExtractedToken string                  // token fetched from the Authorization header
	Parameters     *map[string]interface{} // parsed parameters
	Resource       Resource                // resource data
	Result         HandlerResponse         // resource handler result

	// backend data
	Token interface{}
	User  interface{}
}

// call the request validation methods and the resource function
func (c *Controller) handleRequest(r *Request) Response {

	// generate the logger for this request
	r.Logger = log.New(c.writer, fmt.Sprintf("[%v]%v > ", r.ID, r.Path), log.LstdFlags|log.Lmsgprefix)

	r.Logger.Printf("request recieved [method: %v] [ip: %v]", r.Method, r.IP)

	// set the default contentType
	r.ContentType = "json"

	// handle panic at request operators calls
	defer func() {
		if rcv := recover(); rcv != nil {
			r.Logger.Printf("request operator panicked [err: %v]", rcv)
			r.Result = HandlerResponse{"SE", rcv}
		}
	}()

	// set the request result as OK
	r.Result = HandlerResponse{Code: "OK", Data: Empty}

	// call the request operators
	r.determineAcceptedContentType()
	r.determineResource(&c.routes)
	r.verifyNetwork()
	r.extractAuthorizationToken()
	r.authorizeUser(&c.backend)
	r.parsePayload()
	r.validateResourceParameters(c.validators)
	r.callBackendPreExecution(&c.backend)
	r.callMethod(c.methods)
	r.callBackendPostExecution(&c.backend)

	r.Logger.Printf("finished the request handler job")

	return c.makeResponse(r)
}

// return an HTTP response for the current request result
func (c *Controller) makeResponse(r *Request) Response {
	var err error

	r.Logger.Printf("starting the response assemble [code: %v]", r.Result.Code)

	// check if the response code exists and fetch its data
	code := generalCodes["GEN-0002"]

	if v, ok := c.codes[r.Result.Code]; ok {
		code = v
	} else {
		if v, ok := generalCodes[r.Result.Code]; ok {
			code = v
		}
	}

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
		r.Logger.Println("this response will be returned as XML")
	}

	// assemble the request response with the code and provided data
	response := struct {
		XMLName xml.Name     `json:"-"`
		Meta    ResponseMeta `json:"meta" xml:"meta"`
		Data    interface{}  `json:"data" xml:"data"`
	}{
		XMLName: xml.Name{Local: "response"},
		Data:    r.Result.Data,
		Meta: ResponseMeta{
			ID:      r.ID,
			Time:    time.Now(),
			Code:    r.Result.Code,
			Message: code.Message,
		},
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
		return Response{HTTPCode: generalCodes["GEN-0003"].HTTPCode, Content: []byte{}, Headers: nil}
	}

	r.Logger.Println("API response assembled. returning HTTP response")

	return Response{HTTPCode: code.HTTPCode, Content: content, Headers: headers}
}
