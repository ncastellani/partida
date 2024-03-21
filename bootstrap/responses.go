package bootstrap

import "time"

// Result
// define the response of an API/queue handler function
type Result struct {
	Code    string      // result operation code
	Message string      // custom resource message to return
	Data    interface{} // operation generated data
}

// Code
// define an operation result code, its messages by language and its HTTP code
type Code struct {
	HTTPCode int               `json:"http"`    // HTTP return code
	Message  map[string]string `json:"message"` // messages from the code
}

// Metadata
// define the metadata of the API response
type Metadata struct {
	Code          string            `json:"code"`             // API response code
	Message       map[string]string `json:"message"`          // response code messages
	ID            string            `json:"id"`               // request identifier for debug
	Time          time.Time         `json:"time"`             // datetime of the request answer
	CustomMessage map[string]string `json:"resource_message"` // message returned by the resource
}

// Response
// define the data to be returned to the server
type Response struct {
	HTTPCode int               // HTTP response code of the result
	Content  []byte            // response content
	Headers  map[string]string // response headers
}
