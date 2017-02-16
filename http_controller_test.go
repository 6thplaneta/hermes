package hermes

import (
	"bytes"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

func Test_Http_Controller_Start(t *testing.T) {
	assert.NoError(t, addTempTables())

}
func Test_Http_Create(t *testing.T) {
	cont := NewController(studentColl, "", nil)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/test", func(c *gin.Context) {
		cont.Create(c)
	})

	stu := Student{}
	stu.Title = "mahsa ghoreishi"
	sJson, _ := json.Marshal(stu)
	contentReader := bytes.NewReader(sJson)

	request, _ := http.NewRequest("POST", "/test", contentReader)

	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	assert.Equal(t, "\"Resource created successfully!\"\n", response.Body.String())
	assert.Equal(t, http.StatusCreated, response.Code)
	assert.Equal(t, "1", response.Header().Get("Location"))

	result, err := studentColl.Get(SystemToken, 1, "")
	student := result.(*Student)
	assert.NoError(t, err)
	assert.Equal(t, "mahsa ghoreishi", student.Title)
}

func Test_Http_Update(t *testing.T) {

	cont := NewController(studentColl, "", nil)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.PUT("/test/:id", func(c *gin.Context) {
		cont.Update(c)
	})

	stu := Student{}
	stu.Id = 1
	stu.Title = "updated value"
	stu.Gender_Id = 1
	stu.Sex_Id = 1
	stu.Supervisor_Id = 1

	sJson, _ := json.Marshal(stu)
	contentReader := bytes.NewReader(sJson)

	request, _ := http.NewRequest("PUT", "/test/1", contentReader)

	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	assert.Equal(t, "\"Resource updated successfully!\"\n", response.Body.String())
	assert.Equal(t, http.StatusOK, response.Code)

	result, err := studentColl.Get(SystemToken, 1, "")
	student := result.(*Student)
	assert.NoError(t, err)
	assert.Equal(t, "updated value", student.Title)
}

func Test_Http_Get(t *testing.T) {
	cont := NewController(studentColl, "", application.Cache)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/test/:id", func(c *gin.Context) {
		cont.Get(c)
	})

	request, _ := http.NewRequest("GET", "/test/1", nil)

	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	stu := Student{}
	json.Unmarshal([]byte(response.Body.String()), &stu)
	assert.Equal(t, http.StatusOK, response.Code)
	assert.Equal(t, "updated value", stu.Title)

}

func Test_Http_List(t *testing.T) {
	cont := NewController(studentColl, "", application.Cache)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/test", func(c *gin.Context) {
		cont.List(c)
	})

	request, _ := http.NewRequest("GET", "/test?$page=1&$page_size=1&$search=updated", nil)

	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	stu := []Student{}

	json.Unmarshal([]byte(response.Body.String()), &stu)
	assert.Equal(t, http.StatusOK, response.Code)
	assert.Equal(t, "updated value", stu[0].Title)

	request, _ = http.NewRequest("GET", "/test?$page=1&$page_size=1&$search=test", nil)

	response = httptest.NewRecorder()
	router.ServeHTTP(response, request)
	stu = []Student{}

	json.Unmarshal([]byte(response.Body.String()), &stu)
	assert.Equal(t, http.StatusOK, response.Code)
	assert.Equal(t, 0, len(stu))

	request, _ = http.NewRequest("GET", "/test?$page=1&$page_size=1&title=test", nil)

	response = httptest.NewRecorder()
	router.ServeHTTP(response, request)
	stu = []Student{}

	json.Unmarshal([]byte(response.Body.String()), &stu)
	assert.Equal(t, http.StatusOK, response.Code)
	assert.Equal(t, 0, len(stu))

	//exact filter
	request, _ = http.NewRequest("GET", "/test?$page=1&$page_size=1&title=updated value", nil)

	response = httptest.NewRecorder()
	router.ServeHTTP(response, request)
	stu = []Student{}

	json.Unmarshal([]byte(response.Body.String()), &stu)
	assert.Equal(t, http.StatusOK, response.Code)
	assert.Equal(t, 1, len(stu))
}

func Test_Http_Delete(t *testing.T) {
	cont := NewController(studentColl, "", application.Cache)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.DELETE("/test/:id", func(c *gin.Context) {
		cont.Delete(c)
	})

	request, _ := http.NewRequest("DELETE", "/test/1", nil)

	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)

	assert.Equal(t, "\"Resource deleted successfully!\"\n", response.Body.String())
	assert.Equal(t, http.StatusOK, response.Code)

	_, e := studentColl.Get(SystemToken, 1, "")
	assert.Error(t, e, "sql: no rows in result set")
}

func Test_Http_Controller_End(t *testing.T) {
	assert.NoError(t, rmTempTables())

}
