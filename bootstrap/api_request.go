package bootstrap

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"log"
	"net"
	"reflect"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/ncastellani/partida/utilfunc"
)

// APIRequest
// define an incoming request, with its metadata, payload and useful contents.
type APIRequest struct {
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
	Resource       APIResource             // resource data
	Result         Result                  // resource handler result

	// backend data
	Token interface{}
	User  interface{}
}

// take an pre-assembled request from an API handler
// and call the request functions to perform the desired
// operation and generate an response back to the handler.
func (app *Application) handleAPIRequest(r *APIRequest) APIResponse {

	// generate the logger for this request
	r.Logger = log.New(app.APILogsWriter, fmt.Sprintf("[%v]%v > ", r.ID, r.Path), log.LstdFlags|log.Lmsgprefix)

	r.Logger.Printf("request recieved [method: %v] [ip: %v]", r.Method, r.IP)

	// set the default contentType
	r.ContentType = "json"

	// handle panic at request operators calls
	defer func() {
		if rcv := recover(); rcv != nil {
			r.Logger.Printf("request operator got in panic [err: %v]", rcv)

			r.Result = Result{"SE", rcv}
		}
	}()

	// set the request result as OK
	r.Result = Result{Code: "OK", Data: utilfunc.Empty}

	// call the request operators
	r.determineAcceptedContentType()
	r.determineResource(&app.APIRoutes)
	r.verifyNetwork()
	r.extractAuthorizationToken()
	r.authorizeUser(&app.Backend)
	r.parsePayload()
	r.validateResourceParameters(&app.APIValidators)
	r.callBackendPreExecution(&app.Backend)
	r.callMethod(&app.APIMethods)
	r.callBackendPostExecution(&app.Backend)

	r.Logger.Printf("finished the request handler job")

	return r.makeResponse(app)
}

// return an HTTP response for the current request result.
func (r *APIRequest) makeResponse(app *Application) APIResponse {
	var err error

	r.Logger.Printf("starting the response assemble... [code: %v]", r.Result.Code)

	// check if the response code exists and fetch its data
	code := app.Codes["GEN-0002"]

	if v, ok := app.Codes[r.Result.Code]; ok {
		code = v
	}

	// set the CORS, CACHE and content type headers
	var headers map[string]string = map[string]string{
		"Content-Type":                 "application/json; charset=utf-8",
		"Cache-Control":                "max-age=0,private,must-revalidate,no-cache",
		"Access-Control-Allow-Origin":  "*",
		"Access-Control-Allow-Methods": "*",
		"Access-Control-Allow-Headers": "*",
		"Access-Control-Max-Age":       "86400",
	}

	if r.ContentType == "xml" {
		r.Logger.Println("this response will be returned as XML")

		headers["Content-Type"] = "application/xml"
	}

	// assemble the request response with the code and provided data
	response := struct {
		XMLName xml.Name    `json:"-"`
		Meta    APIMetadata `json:"meta" xml:"meta"`
		Data    interface{} `json:"data" xml:"data"`
	}{
		XMLName: xml.Name{Local: "response"},
		Data:    r.Result.Data,
		Meta: APIMetadata{
			ID:      r.ID,
			Time:    time.Now(),
			Code:    r.Result.Code,
			Message: code.Message,
		},
	}

	// perform the XML/JSON marshaling of the response
	var content []byte

	if r.ContentType == "xml" {
		content, err = xml.MarshalIndent(response, "", "   ")
	} else {
		content, err = json.Marshal(response)
	}

	if err != nil {
		r.Logger.Fatalf("failed to marshal JSON/XML with response [err: %v]", err)

		return APIResponse{HTTPCode: app.Codes["GEN-0003"].HTTPCode, Content: []byte{}, Headers: nil}
	}

	r.Logger.Println("API response assembled. returning HTTP response...")

	return APIResponse{HTTPCode: code.HTTPCode, Content: content, Headers: headers}
}

// update the request result.
func (r *APIRequest) updateResult(code string, data interface{}) {
	r.Result = Result{Code: code, Data: data}

}

// check for content types at the accept header.
func (r *APIRequest) determineAcceptedContentType() {
	if val, ok := r.Headers["Accept"]; ok {
		if val == "application/xml" && r.ContentType == "" {
			r.Logger.Println("application/xml content type found on \"Accept\" header")

			r.ContentType = "xml"
		}
	}

}

// determine the requested route and resource.
func (r *APIRequest) determineResource(routes *map[string]map[string]APIResource) {
	if r.Result.Code != "OK" {
		return
	}

	// check for route existence at the controller
	if _, ok := (*routes)[r.Path]; !ok {
		r.Logger.Printf("route not found [path: %v]", r.Path)

		r.updateResult("GEN-0004", utilfunc.Empty)
		return
	}

	r.Logger.Printf("route exists. checking HTTP method for a resource... [route: %v]", r.Path)

	// check for route methods
	if v, ok := (*routes)[r.Path][r.Method]; !ok {
		r.Logger.Printf("method not available for this route [method: %v]", r.Method)

		// return an OK response for OPTIONS verb validations
		if r.Method == "OPTIONS" {
			r.Logger.Printf("the current request is an OPTIONS check validation")

			r.updateResult("GEN-0005", utilfunc.Empty)
			return
		}

		r.updateResult("GEN-0006", utilfunc.Empty)
		return
	} else {
		r.Resource = v
	}

	r.Logger.Printf("resource exists. a valid HTTP method was used at this route, matching a resource [method: %v]", r.Method)

}

// verify if the network data used by the requester is acceptable for this resource.
func (r *APIRequest) verifyNetwork() {
	if r.Result.Code != "OK" {
		return
	}

	// check if the current IP address pass the network policy
	addrInExceptions := false
	for _, v := range r.Resource.Network.Exceptions {
		_, IPrange, _ := net.ParseCIDR(v)
		userAddr := net.ParseIP(r.IP)
		if !addrInExceptions {
			addrInExceptions = IPrange.Contains(userAddr)
		}
	}

	if (r.Resource.Network.Default == "deny" && !addrInExceptions) || (r.Resource.Network.Default == "allow" && addrInExceptions) {
		r.Logger.Printf("resource does not allow this user IP [ip: %v]", r.IP)

		r.updateResult("GEN-0007", utilfunc.Empty)
		return
	}

	r.Logger.Printf("resource supports the network data of this user [ip: %v]", r.IP)

}

// get the passed user token from the Authorization header.
func (r *APIRequest) extractAuthorizationToken() {
	if r.Result.Code != "OK" || !r.Resource.Authentication {
		return
	}

	r.Logger.Println("trying to fetch auth token from the 'Authorization' header...")

	// check if the header Authorization was passed
	var token string

	if v, ok := r.Headers["Authorization"]; ok {
		token = v
	} else {
		r.Logger.Println("'Authorization' header is not present or does not holds any content")

		r.updateResult("GEN-0008", utilfunc.Empty)
		return
	}

	// get the second element of the authorization header
	authHeader := strings.Fields(token)

	if len(authHeader) == 1 {
		r.Logger.Println("the \"Authorization\" header is present but does not use the correct format")

		r.updateResult("GEN-0009", utilfunc.Empty)
		return
	}

	r.ExtractedToken = authHeader[1]

	r.Logger.Printf("sucessfully obtained the authorization token [token: ...%v]", r.ExtractedToken[len(r.ExtractedToken)-4:])

}

// call the user authorizer at the backend service.
func (r *APIRequest) authorizeUser(be *Backend) {
	if r.Result.Code != "OK" || !r.Resource.Authentication {
		return
	}

	r.Logger.Println("calling the user authorizer at the backend service...")

	r.Result = (*be).APIAuthorizeUser(r)

}

// extract and parse parameters from URL query and body payload.
func (r *APIRequest) parsePayload() {
	if r.Result.Code != "OK" || len(r.Resource.Parameters) == 0 {
		return
	}

	r.Logger.Println("starting the parse of the request payload...")

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
				r.updateResult("GEN-0011", err)
				return
			}
		} else {
			r.updateResult("GEN-0010", r.Headers["Accept"])
			return
		}

		// check if the inputted body is an associative map
		if !utilfunc.IsMap(bodyParameters) {
			r.updateResult("GEN-0012", utilfunc.Empty)
			return
		}

		// determine the keys passed on the body
		for _, v := range reflect.ValueOf(bodyParameters).MapKeys() {
			bodyKeys = append(bodyKeys, v.String())
		}

	}

	// validate the type of each resource parameter
	parameters := make(map[string]interface{})

	var missing []APIResourceParameter
	var invalid []APIResourceParameter

	for _, v := range r.Resource.Parameters {

		// check if the param is on the recieved keys
		var methodParams *map[string]interface{}

		if v.QueryParameter {
			if !utilfunc.StringInSlice(v.Name, queryKeys) {
				r.Logger.Printf("parameter missing at the URL query [param: %v]", v.Name)

				missing = append(missing, v)
				continue
			}

			methodParams = &queryParameters
		} else {
			if !utilfunc.StringInSlice(v.Name, bodyKeys) {
				r.Logger.Printf("parameter missing at the body payload [param: %v]", v.Name)

				missing = append(missing, v)
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
			if !utilfunc.StringInSlice((*methodParams)[v.Name].(string), v.Options) {
				r.Logger.Printf("parameter got an value that does not match the ENUM available ones [param: %v] [recieved: %v]", v.Name, (*methodParams)[v.Name].(string))

				invalid = append(invalid, v)
				continue
			}
		}

		// perform param data check for the "string" type
		if v.Kind == "string" && (*methodParams)[v.Name] != nil {

			// check the length of the recived data
			if v.MaxLength != 0 && (utf8.RuneCountInString((*methodParams)[v.Name].(string)) > v.MaxLength) {
				r.Logger.Printf("parameter surpass the max length for string [param: %v] [maxLength: %v]", v.Name, v.MaxLength)

				invalid = append(invalid, v)
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

		r.updateResult("GEN-0013", struct {
			Missing *[]APIResourceParameter `json:"missing"`
			Invalid *[]APIResourceParameter `json:"invalid"`
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

// call the parameter validator function for each parameter at the resource.
func (r *APIRequest) validateResourceParameters(validators *map[string]APIParameterValidator) {
	if r.Result.Code != "OK" || len(r.Resource.Parameters) == 0 {
		return
	}

	r.Logger.Println("starting the resource parameters validations...")

	// perform the resource params validations
	for _, v := range r.Resource.Parameters {
		for _, validator := range v.Validators {

			// call the parameter validator
			res := (*validators)[validator]((*r.Parameters)[v.Name], r)

			// handle validator errors
			if res.Code != "OK" {
				r.Logger.Printf("parameter validation failed [param: %v] [validator: %v] [returnedCode: %v] [err: %v]", v.Name, validator, res.Code, res.Data)

				r.updateResult(res.Code, struct {
					Parameter     APIResourceParameter `json:"parameter"`
					Input         interface{}          `json:"input"`
					ValidatorData interface{}          `json:"error"`
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

// call the backend pre execution function to perform backend logic.
func (r *APIRequest) callBackendPreExecution(be *Backend) {
	if r.Result.Code != "OK" {
		return
	}

	r.Logger.Println("calling the before-method operation function at backend service...")

	r.Result = (*be).APIBeforeMethodOperations(r)

}

// call the resource method function.
func (r *APIRequest) callMethod(methods *map[string]APIResourceMethod) {
	if r.Result.Code != "OK" {
		return
	}

	// check if the resource method function exists
	if _, ok := (*methods)[r.Resource.ResourceMethod]; !ok {
		r.Logger.Println("resource method function does not exists at the handlers map")

		r.updateResult("GEN-0001", r.Resource.ResourceMethod)
		return
	}

	r.Logger.Println("-- executing resource method --")

	// handle panic at function call
	defer func() {
		if rcv := recover(); rcv != nil {
			r.Logger.Printf("resource method function got in panic [err: %v]", rcv)

			r.Result = Result{"SE", rcv}
		}
	}()

	r.Result = (*methods)[r.Resource.ResourceMethod](r)

	// add the OK code if the return code is empty
	if r.Result.Code == "" {
		r.Result.Code = "OK"
	}

	r.Logger.Println("-- resource method execution ended --")

	r.Logger.Println("sucessfully executed the resource method function")

}

// call the backend post execution function to perform backend logic.
func (r *APIRequest) callBackendPostExecution(be *Backend) {
	if r.Result.Code != "OK" {
		return
	}

	r.Logger.Println("calling the after-method operation function at backend service...")

	r.Result = (*be).APIAfterMethodOperations(r)

}
