package bootstrap

// Backend
// define the interfaces that the client application must implement
// in order to comply with the ways of executing, validating
// and operating that this module do.
type Backend interface {
	APIAuthorizeUser(r *APIRequest) Result
	APIBeforeMethodOperations(r *APIRequest) Result
	APIAfterMethodOperations(r *APIRequest) Result
}
