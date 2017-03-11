package hermes

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"strings"
)

func AuthMiddleware(escapes []string) gin.HandlerFunc {
	return func(c *gin.Context) {

		if authEnabled == false {
			c.Next()
			return
		}

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
