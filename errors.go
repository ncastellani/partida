package partida

import "errors"

var ErrRequestHandlerUnsuccessful = errors.New("[handler] the last validation operation returned an unsuccessful code")
var ErrNoResource = errors.New("[handler] the requested resource does not exist")
