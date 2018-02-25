package hermes

import (
	"net/http"
	"strings"

	"github.com/6thplaneta/u"
	"github.com/gin-gonic/gin"
)

func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Location, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, Cache-Control, X-Requested-With, X-Forwarded-For")
		c.Writer.Header().Set("Access-Control-Expose-Headers", "Content-Type, Location, Authorization, accept, Cache-Control")
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
			// HandleHttpError(c, err, application.Logger)
			return
		}

		r := c.Request

		r.ParseMultipartForm(32 << 20)
		file, handler, err := r.FormFile(inputName)
		if err != nil {
			// HandleHttpError(c, err, application.Logger)
			return
		}
		defer file.Close()
		if fn != nil {
			out, errFN := fn(url, handler.Header.Get("Content-Type"))
			if errFN != nil {
				// HandleHttpError(c, errFN, application.Logger)
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
			// HandleHttpError(c, ErrRateExceed, application.Logger)
		} else {
			c.Next()
		}

	}
}

func LoggerMiddleware(logger *u.Logger2, excludes []string) gin.HandlerFunc {
	return func(c *gin.Context) {
		for i := 0; i < len(excludes); i++ {
			var str string
			str = excludes[i]
			strs := strings.Split(str, ":")
			if c.Request.Method == strs[0] && strings.Contains(c.Request.URL.Path, strs[1]) {
				c.Next()
				return
			}
		}
		//logger.LogHttp(c)
		c.Next()
	}
}
