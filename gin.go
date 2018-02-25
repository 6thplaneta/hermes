package hermes

import (
	"fmt"
	"time"

	"github.com/6thplaneta/go-server/logs"

	"github.com/gin-gonic/gin"
)

func newGinEngine() *gin.Engine {
	engine := gin.New()
	engine.Use(ginLogger(), gin.Recovery())
	return engine
}

func ginLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		if logs.GetLevel() > logs.Trace {
			return
		}

		start := time.Now()
		path := c.Request.URL.Path
		c.Next()
		latency := time.Now().Sub(start)
		clientIP := c.ClientIP()
		method := c.Request.Method
		statusCode := c.Writer.Status()

		// logs.Handle(logs.Trace.NewWithTag("Request", fmt.Sprintf("%3d | %13v | %15s | %-7s | %s",
		logs.Handle(logs.Trace.NewWithTag("Request", fmt.Sprintf("%3d | %v | %s | %s | %s",
			statusCode, latency, clientIP, method, path)))
	}
}
