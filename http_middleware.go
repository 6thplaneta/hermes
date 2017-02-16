package hermes

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Location, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With, X-Forwarded-For")
		c.Writer.Header().Set("Access-Control-Expose-Headers", "Content-Type, Location, Authorization, accept, origin, Cache-Control")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

//
func UploadMiddleware(inputName, savePath string, fn func(string, string) (interface{}, error)) gin.HandlerFunc {
	return func(c *gin.Context) {
		url, err := Upload(c, inputName, savePath)
		if err != nil {
			HandleHttpError(c, err, application.Logger)
			return
		}

		r := c.Request

		r.ParseMultipartForm(32 << 20)
		file, handler, err := r.FormFile(inputName)
		if err != nil {
			HandleHttpError(c, err, application.Logger)
			return
		}
		defer file.Close()
		if fn != nil {
			out, errFN := fn(url, handler.Header.Get("Content-Type"))
			if errFN != nil {
				HandleHttpError(c, errFN, application.Logger)
			} else {
				c.JSON(http.StatusOK, out)
			}
		} else {
			c.JSON(http.StatusOK, url)
		}

	}
}

var GlobalRateLimiter *RateLimiter

func RateLimitMiddleware(rl *RateLimiter) gin.HandlerFunc {
	GlobalRateLimiter = rl
	return func(c *gin.Context) {
		// token := c.Request.Header.Get("Authorization")
		ip := c.Request.RemoteAddr
		pass := rl.Check(ip)
		if !pass {
			HandleHttpError(c, ErrRateExceed, application.Logger)
		} else {
			c.Next()
		}

	}
}

func LoggerMiddleware(logger *Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		logger.LogHttp(c)
		c.Next()

	}
}
