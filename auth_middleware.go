package hermes

import (
	// "fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"strings"
)

func AuthMiddleware(escapes []string) gin.HandlerFunc {
	return func(c *gin.Context) {
		//don't check authentication if auth is disabled
		if authEnabled == false || strings.Contains(c.Request.Method, "OPTIONS") {
			c.Next()
			return
		}
		//don't check auth for the apis do not require authentication
		for i := 0; i < len(escapes); i++ {
			var str string
			str = escapes[i]

			strs := strings.Split(str, ":")
			if c.Request.Method == strs[0] && strings.Contains(c.Request.URL.Path, strs[1]) {
				c.Next()
				return
			}
		}

		token := c.Request.Header.Get("Authorization")

		if token == "" {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		exists, err := AgentTokenColl.Exists(token)
		if err != nil {

			c.JSON(http.StatusInternalServerError, gin.H{"message": Messages["InternalServerError"]})
			return
		}
		if !exists {

			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

	}
}
