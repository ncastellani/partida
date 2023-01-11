package partida

// type for a map of function that return an request response
type ActionHandler func(r *Request) RequestResponse

// type for a map of a function that takes in a param data and validate it
type ActionValidator func(data string, r *Request) (string, int)
