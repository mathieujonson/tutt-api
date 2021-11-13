package main

import (
	"fmt"

	"github.com/gin-gonic/gin"

	slack "tutt-api/slack"
)

func main() {
	fmt.Println("Starting tutt-api...")
	
	r := gin.Default()
	
	r.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "Health check Ok",
		})
	})
	
	r.GET("/api/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "Welcome to the tutt api",
		})
	})

	r.POST("/api/slack/interactive", slack.Interactive)
	
	r.Run() // listen and serve on 0.0.0.0:8080 (for windows "localhost:8080")
}