package hermes

import (
	"fmt"
	"strings"
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
		if logs.GetLevel() < logs.Debug {
			return
		}

		start := time.Now()
		path := c.Request.URL.Path
		c.Next()
		latency := time.Now().Sub(start)
		clientIP := c.ClientIP()
		method := c.Request.Method
		statusCode := c.Writer.Status()
		if logs.GetLevel() == logs.Trace {
			if c.Request.Header.Get("Content-Type") == "application/json" {
				raw, err := c.GetRawData()
				if err != nil {
					println(err.Error())
					return
				}
				body := string(raw)
				body = strings.Replace(body, "\n", "", -1)
				body = strings.Replace(body, "\t", "", -1)
				logs.Handle(logs.Trace.NewWithTag("Middle", fmt.Sprintf("%3d | %v | %s | %s | %s\r\n%s",
					statusCode, latency, clientIP, method, path, string(body))))
			} else {
				logs.Handle(logs.Trace.NewWithTag("Middle", fmt.Sprintf("%3d | %v | %s | %s | %s",
					statusCode, latency, clientIP, method, path)))
			}

			// logs.Handle(logs.Trace.NewWithTag("Request", fmt.Sprintf("%3d | %13v | %15s | %-7s | %s",
		} else {
			// logs.Handle(logs.Trace.NewWithTag("Request", fmt.Sprintf("%3d | %13v | %15s | %-7s | %s",
			logs.Handle(logs.Debug.NewWithTag("Middle", fmt.Sprintf("%3d | %v | %s | %s | %s",
				statusCode, latency, clientIP, method, path)))
		}
	}
}
