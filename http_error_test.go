package hermes

import (
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHandleHttpError(t *testing.T) {
	var err error

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/test", func(c *gin.Context) {
		HandleHttpError(c, err, nil)
	})

	//NotFound
	err = errors.New("NotFound")
	request, _ := http.NewRequest("GET", "/test", nil)
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	assert.Equal(t, "\"Resource not found!\"\n", response.Body.String())
	assert.Equal(t, http.StatusNotFound, response.Code)

	//Forbidden
	err = errors.New("Forbidden")
	request, _ = http.NewRequest("GET", "/test", nil)
	response = httptest.NewRecorder()
	router.ServeHTTP(response, request)
	assert.Equal(t, "\"You are not authorized for access to this resource\"\n", response.Body.String())
	assert.Equal(t, http.StatusForbidden, response.Code)

	//Duplicate data!
	err = errors.New("DuplicateData")
	request, _ = http.NewRequest("GET", "/test", nil)
	response = httptest.NewRecorder()
	router.ServeHTTP(response, request)
	assert.Equal(t, "\"Duplicate data!\"\n", response.Body.String())
	assert.Equal(t, http.StatusConflict, response.Code)

	//server error!
	err = errors.New("error in db example")
	request, _ = http.NewRequest("GET", "/test", nil)
	response = httptest.NewRecorder()
	router.ServeHTTP(response, request)
	assert.Equal(t, "\"error in db example\"\n", response.Body.String())
	assert.Equal(t, http.StatusInternalServerError, response.Code)

	//server error!
	err = &CustomError{"NotValid", "RequiredPassword"}
	request, _ = http.NewRequest("GET", "/test", nil)
	response = httptest.NewRecorder()
	router.ServeHTTP(response, request)
	assert.Equal(t, "\"RequiredPassword\"\n", response.Body.String())
	assert.Equal(t, http.StatusBadRequest, response.Code)
}
