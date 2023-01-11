package partida

import (
	"encoding/base64"
	"io"
	"net/http"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/google/uuid"
)

// handle an inbound HTTP request (via Golang standard HTTP lib)
func (c *Controller) HandlerForHTTP(w http.ResponseWriter, e *http.Request) {

	// get the request input body
	input, _ := io.ReadAll(e.Body)

	// get the IP from request
	ip := strings.Split(e.RemoteAddr, ":")[0]
	if strings.Contains(e.RemoteAddr, "[::1]") {
		ip = "127.0.0.1"
	}

	// iterate over the headers to get the first value
	headers := make(map[string]string)
	for k, v := range e.Header {
		headers[k] = v[0]
	}

	// iterate over the query string params to get the first value
	queryParams := make(map[string]string)
	for k, v := range e.URL.Query() {
		queryParams[k] = v[0]
	}

	// parse the path for getting the action
	path := "index"
	if e.URL.Path != "/" {
		path = e.URL.Path[1:]
	}

	// assemble and perform the request validation and method
	r := Request{
		ID: uuid.New().String(),

		IP:      ip,
		Query:   queryParams,
		Headers: headers,

		method: e.Method,
		path:   path,
		input:  input,
	}

	res := c.handleRequest(&r)

	// set the response headers on the request
	for k, v := range res.Headers {
		w.Header().Set(k, v)
	}

	// return the response to the user
	w.WriteHeader(res.HTTPCode)
	w.Write(res.Content)

	r.Logger.Println("")

}

// handle an inbound AWS Lambda request (via API Gateway v2)
func (c *Controller) HandlerForAWSLambda(e events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {

	// generate the request with the relevant data
	r := Request{
		ID: e.RequestContext.RequestID,

		method:  e.RequestContext.HTTP.Method,
		IP:      e.RequestContext.HTTP.SourceIP,
		Query:   e.QueryStringParameters,
		Headers: e.Headers,
	}

	// parse the path for getting the action
	if e.RawPath != "/" {
		r.path = e.RawPath[1:]
	} else {
		r.path = "index"
	}

	// get the request input body also handling Base64 encoded bodies
	if e.IsBase64Encoded {
		r.input, _ = base64.StdEncoding.DecodeString(e.Body)
	} else {
		r.input = []byte(e.Body)
	}

	// assemble and perform the request validation and method
	res := c.handleRequest(&r)

	r.Logger.Println("")

	return events.APIGatewayV2HTTPResponse{
		StatusCode: res.HTTPCode,
		Headers:    res.Headers,
		Body:       string(res.Content),
	}, nil
}
