package server

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func HelloServer(port string) {
	r := gin.Default()

	r.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"data": "Hello, you are run on port: " + port,
		})
	})
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"data": "healthy",
		})
	})

	r.Run(":" + port)
}
