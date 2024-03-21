package bootstrap

import "time"

// APIResponse
// define the data to be returned to the server
type APIResponse struct {
	HTTPCode int               // HTTP response code of the result
	Content  []byte            // response content
	Headers  map[string]string // response headers
}

// APIMetadata
// define the metadata of the API response
type APIMetadata struct {
	ID      string            `json:"id"`      // request identifier for debug
	Time    time.Time         `json:"time"`    // datetime of the request answer
	Code    string            `json:"code"`    // API response code
	Message map[string]string `json:"message"` // response code messages
}

// APIResource
// define an API method within a route
type APIResource struct {
	ResourceMethod string                 `json:"function"`
	Authentication bool                   `json:"authentication"` // if authentication token should be required
	Network        APIResourceNetwork     `json:"network"`        // network based policies
	Parameters     []APIResourceParameter `json:"parameters"`     // acceptable parameters for this action
}

// APIResourceParameter
// define an parameter specification for the resource
type APIResourceParameter struct {
	Name           string   `json:"name"`            // parameter name
	Kind           string   `json:"kind"`            // parameter type (string/number/enum)
	Required       bool     `json:"required"`        // is required
	MaxLength      int      `json:"max_length"`      // max length of the string (0 for no limit)
	QueryParameter bool     `json:"query_parameter"` // if this parameter should be extracted from the GET query
	Options        []string `json:"options"`         // if type ENUM, this is a list of the available options
	Validators     []string `json:"validators"`      // list of custom functions to validate this parameter
}

// APIResourceNetwork
// default an config to an network restriction
type APIResourceNetwork struct {
	Default    string   `json:"default"`    // default action to perform on network data
	Exceptions []string `json:"exceptions"` // exceptions to the default behavior
}

// APIResourceMethod
// define a correlation (map) between an string and a function
type APIResourceMethod func(r *APIRequest) Result

// APIParameterValidator
// define a correlation (map) between an parameter validator and a function
type APIParameterValidator func(data interface{}, r *APIRequest) Result
