package partida

import (
	"encoding/json"
	"net"
	"reflect"
	"strings"
	"unicode/utf8"

	"gopkg.in/guregu/null.v3"
)

// check if the action exists, HTTP method compatibility and network authorization
func (r *Request) validateRequest(actions map[string]Action) error {

	r.Logger.Println("starting up the general validations procedures")

	// check if the requested API action/version exists
	if v, ok := actions[r.path]; ok {
		r.ActionName = null.NewString(r.path, true)
		r.action = v
	} else {
		r.Logger.Printf("action not found [path: %v] [method: %v]", r.path, r.method)

		return r.updateResult(5, "", Empty, nil)
	}

	r.Logger.Printf("the requested action was selected [action: %v]", r.path)

	// check if the HTTP request method is acceptable by the action
	if r.method == "OPTIONS" {
		r.Logger.Printf("action called for OPTIONS method [action: %v]", r.ActionName.String)

		return r.updateResult(6, "", Empty, nil)
	} else if r.method != r.action.Method {
		r.Logger.Printf("action was called using an unsupported method [action: %v] [method: %v]", r.ActionName.String, r.method)

		return r.updateResult(7, "", Empty, nil)
	}

	r.Logger.Printf("a valid HTTP method for this action was used on this request [method: %v]", r.method)

	// check if the current IP address pass the network policy
	addrInExceptions := false
	for _, v := range r.action.Network.Exception {
		_, IPrange, _ := net.ParseCIDR(v)
		userAddr := net.ParseIP(r.IP)
		if !addrInExceptions {
			addrInExceptions = IPrange.Contains(userAddr)
		}
	}

	if (r.action.Network.Command == "deny" && !addrInExceptions) || (r.action.Network.Command == "allow" && addrInExceptions) {
		r.Logger.Printf("source IP for this request does not have access to this action [action: %v] [ip: %v]", r.ActionName.String, r.IP)

		return r.updateResult(8, "", Empty, nil)
	}

	r.Logger.Printf("this action is able to support the requester IP address [ip: %v]", r.IP)
	r.Logger.Println("finished the general validations successfully")

	return nil
}

// get the passed user token from the Authorization header
func (r *Request) fetchUserToken() error {

	r.Logger.Println("trying to fetch an token from the \"Authorization\" header")

	// check if the header Authorization was passed
	var token string

	if v, ok := r.Headers["Authorization"]; ok {
		token = v
	} else {
		r.Logger.Println("the \"Authorization\" header is not present or does not holds any content")

		return r.updateResult(9, "", Empty, nil)
	}

	// get the second element of the authorization header
	authHeader := strings.Fields(token)
	if len(authHeader) == 1 {
		r.Logger.Println("the \"Authorization\" header is present but does not use the correct format")

		return r.updateResult(9, "handler.invalid_authorization_header", Empty, nil)
	}

	r.HeaderToken = authHeader[1]

	r.Logger.Println("finished the fetch token from header operation")

	return nil
}

// !!
func (r *Request) validateBodyParameters(validators map[string]ActionValidator) error {
	var err error

	r.Logger.Println("starting the body validation procedures for this request")

	// keys and data recieved on input
	var keys []string
	parameters := make(map[string]interface{})

	// get the passed params from each source
	if r.action.Method != "GET" {
		// handle the params passed via ANY method except GET

		r.Logger.Println("the body parameters should be passed as JSON")

		// check if the recieved content is parseable to an inteface
		err = json.Unmarshal(r.input, &parameters)
		if err != nil {
			return r.updateResult(104, "", Empty, nil)
		}

		// check if the request body is an associative JSON
		if !IsAssociative(parameters) {
			return r.updateResult(105, "", Empty, nil)
		}

		// determine the keys passed on the body
		for _, v := range reflect.ValueOf(parameters).MapKeys() {
			keys = append(keys, v.String())
		}

	} else {
		// handle the params passed via GET

		r.Logger.Println("the body parameters should be passed on the URL as query")

		// parse the recieved GET params to the body and keys
		for k, v := range r.Query {
			keys = append(keys, k)
			parameters[k] = v[0]
		}

	}

	// append the general type parameters !!

	// handle the action param validation
	var missing []ActionParam
	var invalid []ActionParam

	for _, v := range r.action.Parameters {

		// check if the param is on the recieved keys
		if !StringInSlice(v.Name, keys) {
			missing = append(missing, v)
			continue
		}

		// check if the informed value is of required type
		switch parameters[v.Name].(type) {
		case string:
			if v.Kind != "string" && v.Kind != "enum" {
				invalid = append(invalid, v)
				continue
			}
		case bool:
			if v.Kind != "bool" {
				invalid = append(invalid, v)
				continue
			}
		case int, int8, int16, int32, int64, float32, float64:
			if v.Kind != "number" {
				invalid = append(invalid, v)
				continue
			}
		case []string, []interface{}:
			if v.Kind != "array" {
				invalid = append(invalid, v)
				continue
			}
		case map[string]interface{}:
			if v.Kind != "map" {
				invalid = append(invalid, v)
				continue
			}
		default:
			if parameters[v.Name] != nil && v.Required {
				invalid = append(invalid, v)
				continue
			}
		}

		// perform param data check for the "enum" type
		if v.Kind == "enum" {
			if !StringInSlice(parameters[v.Name].(string), v.Options) {
				invalid = append(invalid, v)
				continue
			}
		}

		// perform param data check for the "string" type
		if v.Kind == "string" && parameters[v.Name] != nil {

			// check the length of the recived data
			if v.MaxLength != 0 && (utf8.RuneCountInString(parameters[v.Name].(string)) > v.MaxLength) {
				invalid = append(invalid, v)
				continue
			}

		}

	}

	// return the parameters that failed the verification
	if len(invalid) > 0 || len(missing) > 0 {
		r.Logger.Printf("this request has invalid or missing parameters [invalid: %v] [missing: %v]", len(invalid), len(missing))

		return r.updateResult(10, "", struct {
			Missing []ActionParam `json:"missing"`
			Invalid []ActionParam `json:"invalid"`
		}{
			Missing: missing,
			Invalid: invalid,
		}, nil)
	}

	// assign the parsed body on the request
	r.Parameters = parameters

	// perform the action validators validations
	for _, v := range r.action.Parameters {
		for _, validator := range v.Validators {

			// convert the parameter to string
			var paramValue string

			switch parameters[v.Name].(type) {
			case string:
				paramValue = r.Parameters[v.Name].(string)
			case interface{}:
				paramValue = ""
			default:
				paramValue = ""
			}

			// check if the parameter was passed
			if paramValue == "" {
				continue
			}

			// perform the validation
			msg, code := validators[validator](paramValue, r)

			r.Logger.Printf("performed parameter check on action validator [paramName: %v] [validator: %v] [result: %v]", v.Name, validator, code)

			if code != 0 {
				return r.updateResult(code, msg, struct {
					Parameter ActionParam `json:"parameter"`
					Input     string      `json:"input"`
				}{
					Parameter: v,
					Input:     paramValue,
				}, nil)
			}

		}
	}

	r.Logger.Println("successfully validated the passed data for body parameters")

	return nil
}

// execute the action function responsible specified on the actionHandlers variable
func (r *Request) callMethod(handlers map[string]ActionHandler) error {

	// check if the action function exists
	if _, ok := handlers[r.path]; !ok {
		r.Logger.Println("the action method function does not exists at the handlers map (ActionHandler)")

		return r.updateResult(3, "", r.action, nil)
	}

	// handle the panic on the function call
	defer func() {
		if rcv := recover(); rcv != nil {
			r.Logger.Printf("the action method function got into a panic [err: %v]", rcv)

			r.result = RequestResponse{1, "", Empty}
		}
	}()

	r.result = handlers[r.path](r)

	r.Logger.Println("sucessfully executed the action method function")

	return nil
}
