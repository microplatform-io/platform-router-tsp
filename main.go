package main

import (
	"github.com/microplatform-io/platform"
	"github.com/teltechsystems/teaspoon"
	"os"
	"time"
)

var (
	rabbitUser = os.Getenv("RABBITMQ_USER")
	rabbitPass = os.Getenv("RABBITMQ_PASS")
	rabbitAddr = os.Getenv("RABBITMQ_PORT_5672_TCP_ADDR")
	rabbitPort = os.Getenv("RABBITMQ_PORT_5672_TCP_PORT")

	logger = platform.GetLogger("router-tsp")
)

func main() {
	hostname, _ := os.Hostname()

	connMgr := platform.NewAmqpConnectionManager(rabbitUser, rabbitPass, rabbitAddr+":"+rabbitPort, "")

	publisher, err := platform.NewAmqpPublisher(connMgr)
	if err != nil {
		logger.Fatalf("> failed to create publisher: %s", err)
	}

	subscriber, err := platform.NewAmqpSubscriber(connMgr, "router_"+hostname)
	if err != nil {
		logger.Fatalf("> failed to create subscriber: %s", err)
	}

	standardRouter := platform.NewStandardRouter(publisher, subscriber)

	teaspoon.ListenAndServe(":877", teaspoon.HandlerFunc(func(w teaspoon.ResponseWriter, r *teaspoon.Request) {
		routedMessage, err := standardRouter.Route(&platform.RoutedMessage{
			Method:   platform.Int32(int32(r.Method)),
			Resource: platform.Int32(int32(r.Resource)),
			Body:     r.Payload,
		}, 5*time.Second)

		// TODO(bmoyles0117):Don't always assume this is a timeout..
		if err != nil {
			errorBytes, _ := platform.Marshal(&platform.Error{
				Message: platform.String("API Request has timed out"),
			})

			routedMessage = &platform.RoutedMessage{
				Method:   platform.Int32(int32(platform.Method_REPLY)),
				Resource: platform.Int32(int32(platform.Resource_ERROR)),
				Body:     errorBytes,
			}
		}

		w.SetMethod(byte(routedMessage.GetMethod()))
		w.SetResource(int(routedMessage.GetResource()))
		w.Write(routedMessage.GetBody())
	}))
}
