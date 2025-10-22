package main

import (
	"fmt"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}

	return value
}

var serverID = getEnv("SERVER_ID", "server-default")
var port = getEnv("PORT", "8080")
var sessionStore sync.Map

type Session struct {
	UserName  string
	LoginTime time.Time
	ServerID  string
}

type LoginRequest struct {
	UserName string `json:"username"`
	PassWord string `json:"password"`
}

func loginHandler(c *gin.Context) {
	var req LoginRequest

	if err := c.ShouldBindBodyWithJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	sessionID := uuid.New().String()
	session := Session{
		UserName:  req.UserName,
		LoginTime: time.Now(),
		ServerID:  serverID,
	}
	sessionStore.Store(sessionID, session)
	c.SetCookie("session_id", sessionID, 3600, "/", "", false, true)

	c.JSON(http.StatusOK, gin.H{"status": "ok", "data": fmt.Sprintf("get user: %s", req.UserName)})
}

func profileHandler(c *gin.Context) {
	sessionID, err := c.Cookie("session_id")
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	value, ok := sessionStore.Load(sessionID)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Session not found"})
		return
	}

	session := value.(Session)

	c.JSON(http.StatusOK, gin.H{
		"username":   session.UserName,
		"login_time": session.LoginTime,
		"server_id":  session.ServerID,
	})

}

func debugSessionHandler(c *gin.Context) {
	sessions := []map[string]interface{}{}

	sessionStore.Range(func(key, value any) bool {
		session := value.(Session)
		sessions = append(sessions, map[string]interface{}{
			"session_id": key.(string),
			"username":   session.UserName,
			"login_time": session.LoginTime,
			"server_id":  session.ServerID,
		})
		return true
	})

	c.JSON(http.StatusOK, gin.H{
		"server_id":     serverID,
		"session_count": len(sessions),
		"sessions":      sessions,
	})
}

func main() {
	r := gin.Default()
	r.POST("/login", loginHandler)
	r.GET("/profile", profileHandler)
	r.GET("/debug/sessions", debugSessionHandler)
	r.Run(":" + port)
}
