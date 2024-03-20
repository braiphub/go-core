package main

import (
	"context"
	"github.com/braiphub/go-core/log"
	"github.com/braiphub/go-core/trace"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"time"
)

func setupRouter() *gin.Engine {
	r := gin.Default()

	r.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})

	return r
}

func main() {
	r := setupRouter()

	type Object struct {
		Text string
	}
	object := Object{"asdf"}

	logger, _ := log.NewZap("production", 0)

	logger.Debug("webserver is listenning", log.Any("port", 80), log.Any("object", object))
	logger.Info("webserver is listenning", log.Any("port", 80), log.Any("object", object))
	logger.Warn("webserver is listenning", log.Any("port", 80), log.Any("object", object))
	logger.Error("webserver is listenning", errors.New("error message"), log.Any("port", 80), log.Any("object", object))

	ctx := context.Background()
	tracer := trace.NewOpenTelemetry(ctx, "order-service", trace.WithGrpcOltpProvider(ctx, "http://tempo:4317"))
	_, span := tracer.StartSpanWithKind(ctx, trace.KindClient, "root span")
	time.Sleep(time.Millisecond * 30)
	span.Close()

	// listen and serve on 0.0.0.0:8080
	r.Run(":5000")
}
