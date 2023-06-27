package partida

import (
	"encoding/json"
	"net"
	"reflect"
	"strings"
	"unicode/utf8"
)

// update the request result
func (r *Request) updateResult(code, msg string, data interface{}) {
	r.Result = HandlerResponse{Code: code, CustomMessage: msg, Data: data}
}

// check for content types at the accept header
func (r *Request) determineAcceptedContentType() {
	if val, ok := r.Headers["Accept"]; ok {
		if val == "application/xml" && r.ContentType == "" {
			r.Logger.Println("application/xml content type found on \"Accept\" header")
			r.ContentType = "xml"
		}
	}
}

// determine the requested route and resource
func (r *Request) determineResource(routes *map[string]map[string]Resource) {
	if r.Result.Code != "OK" {
		return
	}

	// check for route existence at the controller
	if _, ok := (*routes)[r.Path]; !ok {
		r.Logger.Printf("route not found [path: %v]", r.Path)
		r.updateResult("GEN-0004", "", Empty)
		return
	}

	r.Logger.Printf("route exists. checking HTTP method for a resource [route: %v]", r.Path)

	// check for route methods
	if v, ok := (*routes)[r.Path][r.Method]; !ok {
		r.Logger.Printf("method not available for this route [method: %v]", r.Method)

		// return an OK response for OPTIONS verb validations
		if r.Method == "OPTIONS" {
			r.Logger.Printf("the current request is an OPTIONS check validation")
			r.updateResult("GEN-0005", "", Empty)
			return
		}

		r.updateResult("GEN-0006", "", Empty)
		return
	} else {
		r.Resource = v
	}

	r.Logger.Printf("resource exists. a valid HTTP method was used at this route, matching a resource [method: %v]", r.Method)

}

// verify if the network data used by the requester is acceptable for this resource
func (r *Request) verifyNetwork() {
	if r.Result.Code != "OK" {
		return
	}

	// check if the current IP address pass the network policy
	addrInExceptions := false
	for _, v := range r.Resource.Network.Exception {
		_, IPrange, _ := net.ParseCIDR(v)
		userAddr := net.ParseIP(r.IP)
		if !addrInExceptions {
			addrInExceptions = IPrange.Contains(userAddr)
		}
	}

	if (r.Resource.Network.Default == "deny" && !addrInExceptions) || (r.Resource.Network.Default == "allow" && addrInExceptions) {
		r.Logger.Printf("resource does not allow this user IP [ip: %v]", r.IP)
		r.updateResult("GEN-0007", "", Empty)
		return
	}

	r.Logger.Printf("resource supports the network data of this user [ip: %v]", r.IP)

}

// get the passed user token from the Authorization header
func (r *Request) extractAuthorizationToken() {
	if r.Result.Code != "OK" || !r.Resource.Authentication {
		return
	}

	r.Logger.Println("trying to fetch auth token from the 'Authorization' header")

	// check if the header Authorization was passed
	var token string

	if v, ok := r.Headers["Authorization"]; ok {
		token = v
	} else {
		r.Logger.Println("'Authorization' header is not present or does not holds any content")
		r.updateResult("GEN-0008", "", Empty)
		return
	}

	// get the second element of the authorization header
	authHeader := strings.Fields(token)
	if len(authHeader) == 1 {
		r.Logger.Println("the \"Authorization\" header is present but does not use the correct format")
		r.updateResult("GEN-0009", "", Empty)
		return
	}

	r.ExtractedToken = authHeader[1]

	r.Logger.Printf("sucessfully obtained the authorization token [token: ...%v]", r.ExtractedToken[len(r.ExtractedToken)-4:])

}

// call the user authorizer at the backend service
func (r *Request) authorizeUser(be *Backend) {
	if r.Result.Code != "OK" || !r.Resource.Authentication {
		return
	}

	r.Logger.Println("calling the user authorizer at the backend service")

	r.Result = (*be).PerformUserAuthorization(r)

}

// extract and parse parameters from URL query and body payload
func (r *Request) parsePayload() {
	if r.Result.Code != "OK" || len(r.Resource.Parameters) == 0 {
		return
	}

	r.Logger.Println("starting the parse of the request payload")

	var err error

	// parse the parameters from the URL query
	var queryKeys []string
	queryParameters := make(map[string]interface{})

	for k, v := range r.Query {
		queryKeys = append(queryKeys, k)
		queryParameters[k] = v[0]
	}

	// parse the body parameters
	var bodyKeys []string
	bodyParameters := make(map[string]interface{})

	if len(r.Input) > 0 {
		r.Logger.Println("this request got an body input")

		// parse the input data into an interface
		if r.ContentType == "json" {
			err = json.Unmarshal(r.Input, &bodyParameters)
			if err != nil {
				r.updateResult("GEN-0011", "", err)
				return
			}
		} else {
			r.updateResult("GEN-0010", "", r.Headers["Accept"])
			return
		}

		// check if the inputted body is an associative map
		if !IsAssociative(bodyParameters) {
			r.updateResult("GEN-0012", "", Empty)
			return
		}

		// determine the keys passed on the body
		for _, v := range reflect.ValueOf(bodyParameters).MapKeys() {
			bodyKeys = append(bodyKeys, v.String())
		}

	}

	// validate the type of each resource parameter
	parameters := make(map[string]interface{})
	var missing []ResourceParameter
	var invalid []ResourceParameter

	for _, v := range r.Resource.Parameters {

		// check if the param is on the recieved keys
		var methodParams *map[string]interface{}

		if v.QueryParameter {
			if !StringInSlice(v.Name, queryKeys) {
				missing = append(missing, v)
				r.Logger.Printf("parameter missing at the URL query [param: %v]", v.Name)
				continue
			}

			methodParams = &queryParameters
		} else {
			if !StringInSlice(v.Name, bodyKeys) {
				missing = append(missing, v)
				r.Logger.Printf("parameter missing at the body payload [param: %v]", v.Name)
				continue
			}

			methodParams = &bodyParameters
		}

		// check if the informed value is of required type
		switch (*methodParams)[v.Name].(type) {
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
			if (*methodParams)[v.Name] != nil && v.Required {
				invalid = append(invalid, v)
				continue
			}
		}

		// perform param data check for the "enum" type
		if v.Kind == "enum" {
			if !StringInSlice((*methodParams)[v.Name].(string), v.Options) {
				invalid = append(invalid, v)
				r.Logger.Printf("parameter got an value that does not match the ENUM available ones [param: %v] [recieved: %v]", v.Name, (*methodParams)[v.Name].(string))
				continue
			}
		}

		// perform param data check for the "string" type
		if v.Kind == "string" && (*methodParams)[v.Name] != nil {

			// check the length of the recived data
			if v.MaxLength != 0 && (utf8.RuneCountInString((*methodParams)[v.Name].(string)) > v.MaxLength) {
				invalid = append(invalid, v)
				r.Logger.Printf("parameter surpass the max length for string [param: %v] [maxLength: %v]", v.Name, v.MaxLength)
				continue
			}

		}

		// append this value into the parameters section
		parameters[v.Name] = (*methodParams)[v.Name]

		r.Logger.Printf("sucessfully extracted and parsed parameter [parameter: %v]", v.Name)

	}

	// return the parameters that failed the verification
	if len(invalid) > 0 || len(missing) > 0 {
		r.Logger.Printf("this request has invalid or missing parameters [invalid: %v] [missing: %v]", len(invalid), len(missing))

		r.updateResult("GEN-0013", "", struct {
			Missing *[]ResourceParameter `json:"missing"`
			Invalid *[]ResourceParameter `json:"invalid"`
		}{
			Missing: &missing,
			Invalid: &invalid,
		})

		return
	}

	// assign the parsed body on the request
	r.Parameters = &parameters

	r.Logger.Printf("sucessfully parsed parameters from the URL query and body payload [available: %v]", len(*r.Parameters))

}

// call the parameter validator function for each parameter at the resource
func (r *Request) validateResourceParameters(validators *map[string]ParameterValidator) {
	if r.Result.Code != "OK" || len(r.Resource.Parameters) == 0 {
		return
	}

	r.Logger.Println("starting the resource parameters validations")

	// perform the resource params validations
	for _, v := range r.Resource.Parameters {
		for _, validator := range v.Validators {

			// call the parameter validator
			res := (*validators)[validator]((*r.Parameters)[v.Name], r)

			// handle validator errors
			if res.Code != "OK" {
				r.Logger.Printf("parameter validation failed [param: %v] [validator: %v] [returnedCode: %v] [err: %v]", v.Name, validator, res.Code, res.Data)

				r.updateResult(res.Code, res.CustomMessage, struct {
					Parameter     ResourceParameter `json:"parameter"`
					Input         interface{}       `json:"input"`
					ValidatorData interface{}       `json:"error"`
				}{
					Parameter:     v,
					Input:         (*r.Parameters)[v.Name],
					ValidatorData: res.Data,
				})

				return
			}

			r.Logger.Printf("parameter validation passed [param: %v] [validator: %v]", v.Name, validator)

		}
	}

	r.Logger.Println("successfully validated parameters for this resource")

}

// call the backend pre execution function to perform backend logic
func (r *Request) callBackendPreExecution(be *Backend) {
	if r.Result.Code != "OK" {
		return
	}

	r.Logger.Println("calling the PRE exection operation function at backend service")

	r.Result = (*be).PerformPreExecutionOperations(r)

}

// call the resource method function
func (r *Request) callMethod(methods *map[string]ResourceMethod) {
	if r.Result.Code != "OK" {
		return
	}

	// check if the resource method function exists
	if _, ok := (*methods)[r.Resource.ResourceMethod]; !ok {
		r.Logger.Println("resource method function does not exists at the handlers map")
		r.updateResult("GEN-0001", "", r.Resource.ResourceMethod)
		return
	}

	r.Logger.Println("-- executing resource method --")

	// handle panic at function call
	defer func() {
		if rcv := recover(); rcv != nil {
			r.Logger.Printf("resource method function panicked [err: %v]", rcv)
			r.Result = HandlerResponse{"SE", "", rcv}
		}
	}()

	r.Result = (*methods)[r.Resource.ResourceMethod](r)

	r.Logger.Println("-- resource method execution ended --")

	r.Logger.Println("sucessfully executed the resource method function")

}

// call the backend post execution function to perform backend logic
func (r *Request) callBackendPostExecution(be *Backend) {
	if r.Result.Code != "OK" {
		return
	}

	r.Logger.Println("calling the POST exection operation function at backend service")

	r.Result = (*be).PerformPostExecutionOperations(r)

}
