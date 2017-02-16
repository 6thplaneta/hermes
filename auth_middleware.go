package hermes

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"strings"
)

func AuthMiddleware(escapes []string) gin.HandlerFunc {
	return func(c *gin.Context) {

		for i := 0; i < len(escapes); i++ {

			if strings.Contains(c.Request.URL.Path, escapes[i]) {
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
