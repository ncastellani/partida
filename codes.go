package partida

var generalCodes = map[string]Code{
	"OK": {HTTPCode: 200, Message: map[string]string{
		"en-us": "Operation performed successfully",
	}},
	"SE": {HTTPCode: 500, Message: map[string]string{
		"en-us": "Internal and unknown server error",
	}},

	"GEN-0001": {HTTPCode: 501, Message: map[string]string{
		"en-us": "The requested resource has no implemented function",
	}},
	"GEN-0002": {HTTPCode: 501, Message: map[string]string{
		"en-us": "The code requested by the resource method does not exists. All resource method operations are done",
	}},
	"GEN-0003": {HTTPCode: 500, Message: map[string]string{
		"en-us": "Operations performed sucessfully but an error occured while marshling response",
	}},
	"GEN-0004": {HTTPCode: 404, Message: map[string]string{
		"en-us": "This application is not configured to handle this endpoint",
	}},
	"GEN-0005": {HTTPCode: 202, Message: map[string]string{
		"en-us": "Performing CORS validation for the OPTIONS HTTP method",
	}},
	"GEN-0006": {HTTPCode: 405, Message: map[string]string{
		"en-us": "This API route does not support the used HTTP request method",
	}},
	"GEN-0007": {HTTPCode: 403, Message: map[string]string{
		"en-us": "Your IP address is not allowed to call this API action",
	}},
	"GEN-0008": {HTTPCode: 401, Message: map[string]string{
		"en-us": "The 'Authorization' header required for authenticated actions was not found",
	}},
	"GEN-0009": {HTTPCode: 401, Message: map[string]string{
		"en-us": "The token at the 'Authorization' header must be prefixed by 'Bearer'",
	}},
	"GEN-0010": {HTTPCode: 400, Message: map[string]string{
		"en-us": "The 'Accept' header requested for a different IO method from JSON. This method is currently not supported",
	}},
	"GEN-0011": {HTTPCode: 400, Message: map[string]string{
		"en-us": "Input payload is not a valid JSON",
	}},
	"GEN-0012": {HTTPCode: 400, Message: map[string]string{
		"en-us": "Input payload is not an associative map to interface",
	}},
	"GEN-0013": {HTTPCode: 406, Message: map[string]string{
		"en-us": "Required parameters missing or invalid",
	}},
}
