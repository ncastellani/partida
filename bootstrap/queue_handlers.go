package bootstrap

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/aws/aws-lambda-go/events"
	"github.com/bytedance/sonic"
	"github.com/ncastellani/partida"
)

// QUEUELambdaHandler
// handle an inbound AWS Lambda request
func (app *Application) QUEUELambdaHandler(ctx context.Context, msgs events.SQSEvent) (err error) {

	// setup a new logger
	l := log.New(app.QueueLogsWriter, "", log.Lmsgprefix)

	// handle the messages
	l.Printf("determined the amount of messages [count: %v]", len(msgs.Records))

	for k, msg := range msgs.Records {
		e := QueueEvent{ID: msg.MessageId}

		// generate a logger for this event
		e.Logger = log.New(l.Writer(), fmt.Sprintf("event(%v/%v) [queue_%v] > ", (k+1), len(msgs.Records), partida.RandomString(4)), log.Lmsgprefix)

		e.Logger.Println("======================================================")
		e.Logger.Println("generated a logger for this message")

		// print out the record inputted value
		msgJSON, err := sonic.Marshal(&msg)
		if err != nil {
			return err
		}

		e.Logger.Printf("prepared to handle message [input: %v]", string(msgJSON))

		// determine the event
		if val, ok := msg.MessageAttributes["METHOD"]; !ok {
			e.Logger.Println("no METHOD at the message attributes")
			return errors.New("no METHOD attribute was found")
		} else {
			if val.DataType != "String" {
				e.Logger.Printf("METHOD attribute not at the desired type [type: %v]", val.DataType)
				return errors.New("the passed 'METHOD' data type is not valid (must be string)")
			}
		}

		e.Name = *msg.MessageAttributes["METHOD"].StringValue
		e.Logger.Printf("determined the event name [name: %v]", e.Name)

		// parse the body JSON
		err = sonic.UnmarshalString(msg.Body, &e.Body)
		if err != nil {
			e.Logger.Printf("failed to parse the message JSON body [err: %v]", err)
			return err
		}

		e.Logger.Printf("parsed the message body [body: %v]", e.Body)

		// call the handler
		e.Logger.Println("calling the queue event handler...")

		err = app.callQueueEvent(&e)
		if err != nil {
			e.Logger.Printf("failed to call the queue method [err: %v]", err)
			return err
		}

	}

	return
}

// callQueueEvent
// call the queue method if exists
func (app *Application) callQueueEvent(e *QueueEvent) (err error) {

	// check if the event method exists
	if _, ok := app.QueueMethods[e.Name]; !ok {
		e.Logger.Println("event method function does not exists")
		return fmt.Errorf("event method function does not exists [name: %v]", e.Name)
	}

	e.Logger.Println("-- executing queue method --")

	// handle panic at function call
	defer func() {
		if rcv := recover(); rcv != nil {
			e.Logger.Printf("resource method function panicked [recover: %v]", rcv)
			err = fmt.Errorf("queue method panic [recover: %v]", rcv)
		}
	}()

	err = app.QueueMethods[e.Name](e)

	e.Logger.Println("-- queue method execution ended --")
	e.Logger.Println("sucessfully executed the resource method function")

	return
}
