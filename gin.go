package hermes

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"strings"
	"time"

	"github.com/6thplaneta/go-server/logs"

	"github.com/gin-gonic/gin"
)

func newGinEngine() *gin.Engine {
	engine := gin.New()
	engine.Use(gin.Recovery(), ginLogger())
	return engine
}

func ginLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		if logs.GetLevel() < logs.Debug {
			return
		}

		path := c.Request.URL.RequestURI()

		clientIP := c.ClientIP()
		method := c.Request.Method

		headers := bytes.NewBufferString("")
		for key, value := range c.Request.Header {
			headers.WriteString(fmt.Sprintf("\r\n[%s: %s]", key, strings.Join(value, " | ")))
		}

		var data string
		if headers.Len() != 0 {
			data = fmt.Sprintf("<Headers>%s", headers.String())
		}

		if logs.GetLevel() == logs.Trace && c.Request.Header.Get("Content-Type") == "application/json" {
			raw, err := ioutil.ReadAll(c.Request.Body)
			if err != nil {
				println(err.Error())
				return
			}
			c.Request.Body = &closingBuffer{bytes.NewReader(raw)}
			body := string(raw)
			body = strings.Replace(body, "\n", "", -1)
			body = strings.Replace(body, "\t", "", -1)
			if data != "" {
				data += "\r\n<Body>\r\n" + body
			} else {
				data = "<Body>\r\n" + body
			}
		}

		start := time.Now()
		c.Next()
		latency := time.Now().Sub(start)
		statusCode := c.Writer.Status()

		if logs.GetLevel() == logs.Trace && data != "" {
			logs.Handle(logs.GetLevel().NewWithTag("Request", fmt.Sprintf("%3d | %v | %s | %s | %s\r\n%s",
				statusCode, latency, clientIP, method, path, data)))
		} else {
			logs.Handle(logs.GetLevel().NewWithTag("Request", fmt.Sprintf("%3d | %v | %s | %s | %s",
				statusCode, latency, clientIP, method, path)))
		}
	}
}

type closingBuffer struct {
	*bytes.Reader
}

func (cb *closingBuffer) Close() error {
	return nil
}
