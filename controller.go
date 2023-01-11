package partida

import "io"

type Backend interface {
	PerformUserAuthorization(request *Request) error
	PerformPostExecutionOperations(request *Request) error
}

type Controller struct {
	PathCodes          string
	PathActions        string
	PathCustomMessages string

	writer         io.Writer
	backend        Backend
	handlers       map[string]ActionHandler
	validators     map[string]ActionValidator
	codes          map[int]Code
	actions        map[string]Action
	customMessages map[string]map[string]string
}

func New(bkd Backend, handlers map[string]ActionHandler, validators map[string]ActionValidator, writer io.Writer) *Controller {
	c := &Controller{
		backend:    bkd,
		handlers:   handlers,
		validators: validators,
	}

	return c
}
