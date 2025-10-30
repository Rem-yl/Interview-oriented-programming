package server

import (
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

var PORT string = getEnv("PORT", "8081")

func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}

	return value
}

func HelloServer() {
	r := gin.Default()

	r.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"data": "Hello, you are run on port: " + PORT,
		})
	})

	r.Run(":" + PORT)
}
