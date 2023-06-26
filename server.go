package partida

import (
	"encoding/base64"
	"io"
	"net/http"
	"strings"

	"github.com/aws/aws-lambda-go/events"
)

// HandlerForHTTP handle an inbound HTTP request (via Golang standard HTTP lib)
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
		ID:      RandomString(10),
		IP:      ip,
		Query:   queryParams,
		Headers: headers,
		Method:  e.Method,
		Path:    path,
		Input:   input,
	}

	res := c.handleRequest(&r)

	// append the request ID
	res.Headers["x-request-id"] = r.ID

	// set the response headers on the request
	for k, v := range res.Headers {
		w.Header().Set(k, v)
	}

	// return the response to the user
	w.WriteHeader(res.HTTPCode)
	w.Write(res.Content)

	r.Logger.Println("DONE")

}

// HandlerForAWSLambda handle an inbound AWS Lambda request (via API Gateway v2)
func (c *Controller) HandlerForAWSLambda(e events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {

	// generate the request with the relevant data
	r := Request{
		ID:      e.RequestContext.RequestID,
		IP:      e.RequestContext.HTTP.SourceIP,
		Method:  e.RequestContext.HTTP.Method,
		Query:   e.QueryStringParameters,
		Headers: e.Headers,
	}

	// parse the path for getting the action
	r.Path = "index"

	if e.RawPath != "/" {
		r.Path = e.RawPath[1:]
	}

	// get the request input body also handling Base64 encoded bodies
	if e.IsBase64Encoded {
		r.Input, _ = base64.StdEncoding.DecodeString(e.Body)
	} else {
		r.Input = []byte(e.Body)
	}

	// assemble and perform the request validation and method
	res := c.handleRequest(&r)

	r.Logger.Println("DONE")

	// append the request ID
	res.Headers["x-request-id"] = r.ID

	return events.APIGatewayV2HTTPResponse{
		StatusCode: res.HTTPCode,
		Headers:    res.Headers,
		Body:       string(res.Content),
	}, nil
}
