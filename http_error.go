package hermes

import (
	"net/http"
	"strings"

	"github.com/6thplaneta/go-server/logs"

	"github.com/gin-gonic/gin"
)

// HandleHttpError gets error key and returns appropriate message regarding to this key
func HandleHttpError(c *gin.Context, err error) {
	// var txt string
	txt := "HTTP Request, Method: " + c.Request.Method + " IP: " + c.ClientIP() + " Path:" + c.Request.RequestURI
	serverName, serverIp, err1 := HostInfo()
	if err1 == nil {
		txt = serverName + " " + serverIp + " "
	}
	txt += c.Request.RequestURI + " " + c.Request.Method + " " + c.ClientIP()
	if logs.GetLevel() >= logs.Debug {
		logs.Handle(logs.Trace.NewWithTag("Request", txt+err.Error()))
	}
	//if logger.Level >= 5 {
	//	token := c.Request.Header.Get("Authorization")
	//	if token == "" {
	//		txt = txt + "empty "
	//
	//	} else {
	//		txt = txt + token + " "
	//	}
	//
	//}

	// if logger != nil {
	// 	logger.Error(txt + err.Error())
	// }
	// all errors are internal unless equal specified errors or have Error structure and NotValid/BadRequest Key
	statusCode := http.StatusInternalServerError
	if err == ErrNotFound {

		statusCode = http.StatusNotFound
	} else if err == ErrForbidden || err == ErrTokenInvalid || err == ErrAgentNotActive {
		statusCode = http.StatusForbidden
	} else if err == ErrObjectInvalid || err == ErrPassRequired || err == ErrPassword ||
		err == ErrPassFormat || err == ErrRateExceed {
		statusCode = http.StatusBadRequest
	} else if err == ErrDuplicate || strings.Contains(err.Error(), Messages["DuplicateIndex"]) {
		statusCode = http.StatusConflict
	}

	if customError, ok := err.(Error); ok {
		if customError.Key == "NotValid" || customError.Key == "BadRequest" {
			statusCode = http.StatusBadRequest
		}
	}

	c.JSON(statusCode, err.Error())
	// c.Abort()
}
