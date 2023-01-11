package partida

import (
	"io"
)

type Backend interface {
	PerformUserAuthorization(request *Request) error
	PerformPostExecutionOperations(request *Request) error
}

type Controller struct {
	writer     io.Writer
	backend    Backend
	handlers   map[string]ActionHandler
	validators map[string]ActionValidator
	codes      map[int]Code
	actions    map[string]Action
	messages   map[string]map[string]string
}

// create a new Controller based on the passed data
func New(
	bkd Backend,
	handlers map[string]ActionHandler,
	validators map[string]ActionValidator,
	writer io.Writer,
	actionsPath string,
	codesPath string,
	messagesPath string,
) *Controller {
	c := &Controller{
		writer:     writer,
		backend:    bkd,
		handlers:   handlers,
		validators: validators,
	}

	ParseJSON(actionsPath, &c.actions)
	ParseJSON(codesPath, &c.codes)
	ParseJSON(messagesPath, &c.messages)

	return c
}
