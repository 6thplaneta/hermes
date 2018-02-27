package hermes

import (
	"bytes"
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
		path := c.Request.URL.RequestURI()
		c.Next()
		latency := time.Now().Sub(start)
		clientIP := c.ClientIP()
		method := c.Request.Method
		statusCode := c.Writer.Status()

		headers := bytes.NewBufferString("")
		for key, value := range c.Request.Header {
			headers.WriteString(fmt.Sprintf("\r\n[%s: %s]", key, strings.Join(value, " | ")))
		}

		var data string
		if headers.Len() != 0 {
			data = fmt.Sprintf("<Headers>%s", headers.String())
		}

		if logs.GetLevel() == logs.Trace && c.Request.Header.Get("Content-Type") == "application/json" {
			raw, err := c.GetRawData()
			if err != nil {
				println(err.Error())
				return
			}
			body := string(raw)
			body = strings.Replace(body, "\n", "", -1)
			body = strings.Replace(body, "\t", "", -1)
			if data != "" {
				data += "\r\n<Body>\r\n" + body
			} else {
				data = "<Body>\r\n" + body
			}
		}

		if logs.GetLevel() == logs.Trace && data != "" {
			logs.Handle(logs.GetLevel().NewWithTag("Request", fmt.Sprintf("%3d | %v | %s | %s | %s\r\n%s",
				statusCode, latency, clientIP, method, path, data)))
		} else {
			logs.Handle(logs.GetLevel().NewWithTag("Request", fmt.Sprintf("%3d | %v | %s | %s | %s",
				statusCode, latency, clientIP, method, path)))
		}
	}
}
