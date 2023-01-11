package partida

// application default response messages and HTTP codes
type Code struct {
	HTTPCode int               `json:"http"`    // HTTP return code
	Message  map[string]string `json:"message"` // messages from the code
}
