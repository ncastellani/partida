package bootstrap

import "log"

// - api

// Resource
// define an API method within a route
type Resource struct {
	ResourceMethod string              `json:"function"`
	Authentication bool                `json:"authentication"` // if authentication token should be required
	Network        ResourceNetwork     `json:"network"`        // network based policies
	Parameters     []ResourceParameter `json:"parameters"`     // acceptable parameters for this action
}

// ResourceParameter
// define an parameter specification for the resource
type ResourceParameter struct {
	Name           string   `json:"name"`            // parameter name
	Kind           string   `json:"kind"`            // parameter type (string/number/enum)
	Required       bool     `json:"required"`        // is required
	MaxLength      int      `json:"max_length"`      // max length of the string (0 for no limit)
	QueryParameter bool     `json:"query_parameter"` // if this parameter should be extracted from the GET query
	Options        []string `json:"options"`         // if type ENUM, this is a list of the available options
	Validators     []string `json:"validators"`      // list of custom functions to validate this parameter
}

// ResourceNetwork
// default an config to an network restriction
type ResourceNetwork struct {
	Default    string   `json:"default"`    // default action to perform on network data
	Exceptions []string `json:"exceptions"` // exceptions to the default behavior
}

// ResourceMethod
// define a correlation (map) between an string and a function
type ResourceMethod func(r *Request) Result

// ParameterValidator
// define a correlation (map) between an parameter validator and a function
type ParameterValidator func(data interface{}, r *Request) Result

// - queue

// !!
type QueueEvent struct {
	Name    string
	ID      string
	Logger  *log.Logger
	Account int
	Body    map[string]interface{}
}

// !!
type QueueMethod func(r *QueueEvent) error
