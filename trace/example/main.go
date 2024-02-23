package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"time"

	"github.com/braiphub/go-core/trace"
	"github.com/bxcodec/faker/v4"
)

func main() {
	if err := run(); err != nil {
		log.Fatalln(err)
	}
}

func run() (err error) {
	// Handle SIGINT (CTRL+C) gracefully.
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, os.Kill)
	defer stop()

	// Set up OpenTelemetry.
	tracer := trace.NewOpenTelemetry(
		ctx,
		"ms-orders",
		trace.WithGrpcOltpProvider(ctx, "http://localhost:4317"),
		// trace.WithHttpOltpProvider(ctx, "http://localhost:4318"),
	)
	defer tracer.Close(ctx)

	latencyRnd := func() time.Duration {
		latency, err := faker.RandomInt(1, 10, 100)
		if err != nil {
			return 10
		}

		return time.Duration(latency[0])
	}

	spanCtx, span1 := tracer.StartSpanWithKind(context.Background(), trace.KindServer, "span")
	span1.Status(trace.StatusError, "err message")
	time.Sleep(latencyRnd() * time.Millisecond)

	_, subSpan := tracer.StartSpan(spanCtx, "span.sub-span", trace.Attr("attr-str", "val"), trace.Attr("attr-int", 2), trace.Attr("attr-bool", false))
	subSpan.Status(trace.StatusOK, "")
	subSpan.Close()

	time.Sleep(latencyRnd() * time.Millisecond)
	span1.Close()

	println("running")
	<-ctx.Done()

	println("closing")
	// const waitSecondsBeforeClose = 10
	// time.Sleep(time.Second * waitSecondsBeforeClose)

	return
}

// func newHTTPHandler() http.Handler {
// 	mux := http.NewServeMux()

// 	// handleFunc is a replacement for mux.HandleFunc
// 	// which enriches the handler's HTTP instrumentation with the pattern as the http.route.
// 	handleFunc := func(pattern string, handlerFunc func(http.ResponseWriter, *http.Request)) {
// 		// Configure the "http.route" for the HTTP instrumentation.
// 		handler := otelhttp.WithRouteTag(pattern, http.HandlerFunc(handlerFunc))
// 		mux.Handle(pattern, handler)
// 	}

// 	// Register handlers.
// 	handleFunc("/rolldice", func(w http.ResponseWriter, r *http.Request) {
// 		roll := 1 + rand.Intn(6)

// 		resp := strconv.Itoa(roll) + "\n"
// 		if _, err := io.WriteString(w, resp); err != nil {
// 			log.Printf("Write failed: %v\n", err)
// 		}
// 	})

// 	// Add HTTP instrumentation for the whole server.
// 	handler := otelhttp.NewHandler(mux, "/")
// 	return handler
// }
