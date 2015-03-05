package main

import (
	"github.com/microplatform-io/platform"
	"github.com/teltechsystems/teaspoon"
	"os"
	"time"
)

func main() {
	hostname, _ := os.Hostname()

	standardRouter := platform.NewStandardRouter(platform.GetDefaultPublisher(), platform.GetDefaultConsumerFactory("router_"+hostname))

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

		w.SetResource(int(routedMessage.GetResource()))
		w.Write(routedMessage.GetBody())
	}))
}
