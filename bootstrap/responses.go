package bootstrap

// Result
// define the response of an API/queue handler function
type Result struct {
	Code string      // result operation code
	Data interface{} // operation generated data
}

// Code
// define an operation result code, its messages by language and its HTTP code
type Code struct {
	HTTPCode int               `json:"http"`    // HTTP return code
	Message  map[string]string `json:"message"` // messages from the code
}
