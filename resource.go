package partida

// Resource is an API method within a route
type Resource struct {
	ResourceMethod string              `json:"function"`
	Authentication bool                `json:"authentication"` // if authentication token should be required
	Network        ResourceNetwork     `json:"network"`        // network based policies
	Parameters     []ResourceParameter `json:"parameters"`     // acceptable parameters for this action
}

// ResourceParameter is an parameter spec for the resource item
type ResourceParameter struct {
	QueryParameter bool     `json:"query_parameter"`
	Name           string   `json:"name"`       // parameter name
	Kind           string   `json:"kind"`       // parameter type (string/number/enum)
	Required       bool     `json:"required"`   // is required
	MaxLength      int      `json:"max_length"` // max length of the string (0 for none)
	Options        []string `json:"options"`    // if type enum, what are the options?

	Validators []string `json:"validators"` // custom functions to validate a parameter
}

// ResourceNetwork is an config to an network restriction
type ResourceNetwork struct {
	Default   string   `json:"default"`   // default action to perform on network data
	Exception []string `json:"exception"` // exceptions to the default behavior
}

// ResourceMethod is correlation (map) between an string and a function
type ResourceMethod func(r *Request) HandlerResponse

// ParameterValidator is a correlation (map) between an parameter validator and a function
type ParameterValidator func(data interface{}, r *Request) HandlerResponse
