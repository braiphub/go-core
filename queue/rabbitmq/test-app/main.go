package main

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/braiphub/go-core/queue"
	"github.com/braiphub/go-core/queue/rabbitmq"
)

func main() {
	const user = "guest"
	const pass = "guest"
	const host = "localhost"
	const listenQueue = "test-queue"
	const dlxExchange = "test-exchange.dlx"
	const publishExchange = "test-exchange"
	const paralelPublishers = 10

	fmt.Println("---------------------------------------------")
	fmt.Println("Press the Enter Key to stop anytime")
	fmt.Println("---------------------------------------------")

	q, err := rabbitmq.New(user, pass, host)
	if err != nil {
		panic(err)
	}
	start := time.Now()

	// process received messages
	totalReceived := 0
	go func(q queue.QueueI) {
		q.Subscribe(context.Background(), listenQueue, dlxExchange, func(m queue.Message) error {
			// insert your business logic here
			totalReceived++
			if totalReceived%2 == 0 {
				// println("message moved do dlx")
				return errors.New("unknown")
			}

			// println("consumed message")
			return nil
		})
	}(q)

	// publish some messages
	totalPublished := 0
	for i := 0; i < paralelPublishers; i++ {
		go func() {
			for {
				q.Publish(
					context.Background(),
					publishExchange,
					queue.Message{
						Metadata: map[string]string{
							"key1":  "val1",
							"index": strconv.Itoa(totalPublished),
						},
						Body: []byte("message payload"),
					},
				)

				totalPublished++
			}
		}()
	}

	fmt.Scanln()
	fmt.Printf("Published messages: %d\n", totalReceived)
	fmt.Printf("Processed messages: %d\n", totalReceived)
	fmt.Printf("Time elapsed (secs): %s\n", time.Since(start).String())
}
