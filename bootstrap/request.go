package bootstrap

import "log"

// Request
// define an incoming request, with its metadata, payload and useful contents
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
	Result         Result                  // resource handler result

	// backend data
	Token interface{}
	User  interface{}
}
