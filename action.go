package partida

// set of information to define an API action
type Action struct {
	Method         string        `json:"method"`         // allowed HTTP method
	Authentication bool          `json:"authentication"` // if authentication token should be required
	Network        ActionNetwork `json:"network"`        // network based policies
	Parameters     []ActionParam `json:"parameters"`     // acceptable parameters for this action

	ShowUserEvent bool `json:"show_user_event"` // if this request should be visible to user on its events
}

// conditions to match the action requirements with the request input
type ActionNetwork struct {
	Command   string   `json:"command"`   // default action to perform on network data
	Exception []string `json:"exception"` // exceptions to the default behavior
}

// define the information of a parameter used by an action
type ActionParam struct {
	Name       string   `json:"name"`       // parameter name
	Kind       string   `json:"kind"`       // parameter type (string/int/enum)
	Validators []string `json:"validators"` // custom functions to validate a parameter
	Required   bool     `json:"required"`   // is required
	MaxLength  int      `json:"max_length"` // max length of the string (0 for none)
	Options    []string `json:"options"`    // if type enum, what are the options?
}
