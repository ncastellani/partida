package partida

import (
	"io"
	"log"
)

// Backend is an implementable interface for servers
// to perform unique business logic on certain operations
type Backend interface {
	PerformUserAuthorization(request *Request) HandlerResponse
	PerformPreExecutionOperations(request *Request) HandlerResponse
	PerformPostExecutionOperations(request *Request) HandlerResponse
}

type Controller struct {
	backend Backend   // backend service
	writer  io.Writer // logger writer

	// responses and routes
	codes    map[string]Code
	routes   map[string]map[string]Resource
	messages map[string]map[string]string

	// function methods for resources and param validators
	validators *map[string]ParameterValidator
	methods    *map[string]ResourceMethod
}

func NewController(bkd Backend) *Controller {
	return &Controller{
		backend: bkd,
		writer:  log.Default().Writer(),
	}
}

func (c *Controller) ParseBackendConfigs(codes, routes, messages string) {
	ParseJSON(codes, &c.codes)
	ParseJSON(routes, &c.routes)
	ParseJSON(messages, &c.messages)
}

func (c *Controller) SetMethods(validators *map[string]ParameterValidator, methods *map[string]ResourceMethod) {
	c.validators = validators
	c.methods = methods
}

func (c *Controller) SetWriter(writer io.Writer) {
	c.writer = writer
}
