package partida

var generalCodes = map[string]Code{
	"OK": {HTTPCode: 200, Message: map[string]string{
		"en-us": "Operation performed successfully",
	}},

	"GN1": {HTTPCode: 500, Message: map[string]string{
		"en-us": "Internal and unknown server error",
	}},
	"GN2": {HTTPCode: 501, Message: map[string]string{
		"en-us": "The requested resource has no implemented function",
	}},
	"GN3": {HTTPCode: 501, Message: map[string]string{
		"en-us": "The code requested by the resource method does not exists. All resource method operations are done",
	}},
	"GN4": {HTTPCode: 500, Message: map[string]string{
		"en-us": "Operations performed sucessfully but an error occured while marshling response",
	}},
	"GN100": {HTTPCode: 404, Message: map[string]string{
		"en-us": "This application is not configured to handle this endpoint",
	}},
	"GN101": {HTTPCode: 206, Message: map[string]string{
		"en-us": "Performing CORS validation for the OPTIONS HTTP method",
	}},
	"GN102": {HTTPCode: 405, Message: map[string]string{
		"en-us": "This API route does not support the used HTTP request method",
	}},
	"GN103": {HTTPCode: 403, Message: map[string]string{
		"en-us": "Your IP address is not allowed to call this API action",
	}},
	"GN104": {HTTPCode: 401, Message: map[string]string{
		"en-us": "The 'Authorization' header required for authenticated actions was not found",
	}},
	"GN105": {HTTPCode: 401, Message: map[string]string{
		"en-us": "The token at the 'Authorization' header must be prefixed by 'Bearer'",
	}},
	"GN106": {HTTPCode: 400, Message: map[string]string{
		"en-us": "The 'Accept' header requested for a different IO method from JSON. This method is currently not supported",
	}},
	"GN107": {HTTPCode: 400, Message: map[string]string{
		"en-us": "Input payload is not a valid JSON",
	}},
	"GN108": {HTTPCode: 400, Message: map[string]string{
		"en-us": "Input payload is not an associative map to interface",
	}},
	"GN109": {HTTPCode: 406, Message: map[string]string{
		"en-us": "Required parameters missing or invalid",
	}},
}
