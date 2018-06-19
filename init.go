package hermes

import (
	"github.com/gin-gonic/gin"
	uuid "github.com/satori/go.uuid"
)

func init() {
	code, _ := uuid.NewV4()
	SystemToken = code.String()
	gin.SetMode(gin.ReleaseMode)
	InitMessages()
}
