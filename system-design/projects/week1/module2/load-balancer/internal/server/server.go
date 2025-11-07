package server

import (
	"math/rand"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// simulateServerDelayNormal 模拟正态分布延迟
// meanMs: 平均耗时，stdDevMs: 标准差（波动）
func simulateServerDelayNormal(meanMs, stdDevMs float64) {
	delay := rand.NormFloat64()*stdDevMs + meanMs
	if delay < 0 {
		delay = 0
	}
	time.Sleep(time.Duration(delay) * time.Millisecond)
}

func HelloServer(port string) {
	r := gin.Default()

	r.GET("/", func(c *gin.Context) {
		simulateServerDelayNormal(20, 10)
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
