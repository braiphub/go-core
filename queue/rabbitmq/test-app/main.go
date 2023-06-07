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
	const vHost = "braip"
	const paralelPublishers = 1

	// without routing key
	// const listenQueue = "test-queue"
	// const dlxExchange = "test-exchange.dlx"
	// const publishExchange = "test-exchange"
	// routingKeys := []string{""}

	// with routing key
	const listenQueue = "router.app2"
	const dlxExchange = "test-exchange.dlx"
	const publishExchange = "router"
	routingKeys := []string{"app2"}

	fmt.Println("---------------------------------------------")
	fmt.Println("Press the Enter Key to stop anytime")
	fmt.Println("---------------------------------------------")

	q, err := rabbitmq.New(user, pass, host, vHost)
	if err != nil {
		panic(err)
	}
	start := time.Now()

	// process received messages
	totalReceived := 0
	go q.Subscribe(context.Background(), listenQueue, dlxExchange, func(_ context.Context, m queue.Message) error {
		// insert your business logic here
		fmt.Printf("received message: buf: %s || metadata: %v\n", m.Body, m.Metadata)

		// 50% ok / 50% dlx
		totalReceived++
		if totalReceived%2 == 0 {
			// println("message moved do dlx")
			return errors.New("unknown")
		}

		return nil
	})

	// publish some messages
	totalPublished := 0
	for i := 0; i < paralelPublishers; i++ {
		go func() {
			for {
				q.Publish(
					context.Background(),
					publishExchange,
					routingKeys,
					queue.Message{
						Metadata: map[string]string{
							"key1":  "val1",
							"index": strconv.Itoa(totalPublished),
						},
						Body: []byte(fmt.Sprintln("message payload", totalPublished)),
					},
				)

				totalPublished++
			}
		}()
	}

	fmt.Scanln()
	fmt.Printf("Published messages: %d\n", totalPublished)
	fmt.Printf("Processed messages: %d\n", totalReceived)
	fmt.Printf("Time elapsed (secs): %s\n", time.Since(start).String())
}
