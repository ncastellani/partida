package bootstrap

import "log"

// !!
type QueueEvent struct {
	Name   string
	ID     string
	Logger *log.Logger
	Body   map[string]interface{}
}

// !!
type QueueMethod func(r *QueueEvent) error
