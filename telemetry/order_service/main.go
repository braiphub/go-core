package main

import "github.com/gin-gonic/gin"

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

	// listen and serve on 0.0.0.0:8080
	println("running")
	r.Run(":5000")
}
