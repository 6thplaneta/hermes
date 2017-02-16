package hermes

import (
	// "errors"
	// "fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"net/url"
	"reflect"
	"strconv"
	"strings"
)

type ControllerInfo struct {
	Name string `json:"name"`
	Path string `json:"path"`
	Show bool   `json:"show"`
}

type Controller struct {
	Coll     Collectionist
	Instance interface{}
	Typ      reflect.Type
	Base     string
}

type Controlist interface {
	List(*gin.Context)
	Get(*gin.Context)
	Create(*gin.Context)
	Delete(*gin.Context)
	Update(*gin.Context)
	Report(*gin.Context)
	Rel(*gin.Context)
	GetRel(*gin.Context)
	UnRel(*gin.Context)
	UpdateRel(*gin.Context)
	GetCollection() Collectionist
	Meta(*gin.Context)
	GetBase() string
	SetBase(string)
	GetInfo() ControllerInfo
	ReadParams(*gin.Context) *Params
}

func (cont *Controller) GetInfo() ControllerInfo {
	ci := ControllerInfo{}
	tp := cont.Coll.GetInstanceType()
	ci.Name = tp.Name()
	ci.Path = cont.GetBase()
	tpID, _ := tp.FieldByName("Id")
	hermesStr := tpID.Tag.Get("hermes")
	ui_html := GetValueOfTagByKey(hermesStr, "ui-html")
	ci.Show = true
	if ui_html == "None" {
		ci.Show = false
	}
	return ci
}
func (cont *Controller) GetCollection() Collectionist {
	return cont.Coll
}

func NewController(coll Collectionist, base string) *Controller {
	inst := coll.GetInstance()
	typ := coll.GetInstanceType()
	cont := &Controller{coll, inst, typ, base}
	return cont
}

func getAggregation(vals url.Values) map[string][]string {

	keywords := []string{"$sum", "$avg", "$count", "$group_by"}

	mymap := make(map[string][]string)

	for key, value := range vals {
		for _, keyword := range keywords {
			if keyword == key {
				mymap[key[1:]] = strings.Split(value[0], ",")
			}
		}
	}
	return mymap
}

func (cont *Controller) Report(c *gin.Context) {
	page, _ := strconv.Atoi(c.Query("$page"))
	pageSize, _ := strconv.Atoi(c.Query("$page_size"))
	if pageSize == 0 {
		pageSize = 10
	}
	sortBy := c.Query("$sort_by")
	sortOrder := c.Query("$sort_order")

	// var user User
	params := cont.ReadParams(c)

	aggregation := getAggregation(c.Request.URL.Query())
	token := c.Request.Header.Get("Authorization")
	result, err := cont.Coll.Report(token, page, pageSize, params, c.Query("$search"), sortBy, sortOrder, c.Query("$populate"), aggregation)
	if err != nil {
		HandleHttpError(c, err, application.Logger)
		return
	}
	c.JSON(http.StatusOK, result)
}

func (cont *Controller) Meta(c *gin.Context) {

	meta := cont.Coll.Meta()
	c.JSON(http.StatusOK, meta)
}

func RedirectPage(c *gin.Context) bool {

	if c.Query("$page") == "" {
		qr := c.Request.URL.Query()
		qr.Set("$page", "1")
		c.Request.URL.RawQuery = qr.Encode()
		if GlobalRateLimiter != nil {
			GlobalRateLimiter.Reset(c.Request.RemoteAddr)
		}
		c.Redirect(http.StatusMovedPermanently, c.Request.URL.String())
		return true
	}
	return false
}

func (cont *Controller) List(c *gin.Context) {
	// if RedirectPage(c) {
	// 	return
	// }

	params := cont.ReadParams(c)
	// var pageInfo PageInfo
	token := c.Request.Header.Get("Authorization")
	pg, err := ReadPaging(c)
	if err != nil {
		HandleHttpError(c, err, application.Logger)
		return
	}

	result, err := CacheList(cont.Coll.GetDataSrc(), cont.Coll, token, params, pg, c.Query("$populate"), "")
	// result, err := cont.Coll.List(token, params, pg, c.Query("$populate"), "")

	if err != nil {
		HandleHttpError(c, err, application.Logger)
		return
	}
	c.JSON(http.StatusOK, result)

}

func (cont *Controller) Get(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	token := c.Request.Header.Get("Authorization")

	// result, err := cont.Coll.Get(token, id, c.Query("$populate"))
	result, err := CacheGet(cont.Coll.GetDataSrc(), cont.Coll, token, id, c.Query("$populate"))
	if err != nil {
		HandleHttpError(c, err, application.Logger)
		return
	}
	c.JSON(http.StatusOK, result)

}

func (cont *Controller) Delete(c *gin.Context) {
	id, iderr := strconv.Atoi(c.Param("id"))
	if iderr != nil {
		nferr := ErrNotFound
		HandleHttpError(c, nferr, application.Logger)
		return
	}
	token := c.Request.Header.Get("Authorization")
	if err := cont.Coll.Delete(token, id); err != nil {
		HandleHttpError(c, err, application.Logger)
		return
	}

	c.JSON(http.StatusOK, Messages["ResourceDeleted"])

}

func (cont *Controller) Update(c *gin.Context) {
	inst := reflect.New(cont.Typ)
	obj := inst.Interface()
	id, iderr := strconv.Atoi(c.Param("id"))
	if iderr != nil {
		nferr := ErrNotFound
		HandleHttpError(c, nferr, application.Logger)
		return
	}

	c.BindJSON(obj)
	token := c.Request.Header.Get("Authorization")
	err := cont.Coll.Update(token, id, obj)
	if err != nil {
		HandleHttpError(c, err, application.Logger)
		return
	}

	c.JSON(http.StatusOK, Messages["ResourceUpdated"])
}

func (cont *Controller) Create(c *gin.Context) {
	inst := reflect.New(cont.Typ)
	obj := inst.Interface()
	err := c.BindJSON(obj)
	if err != nil {
		HandleHttpError(c, NewError("BadRequest", "Error in parsing body: "+err.Error()), application.Logger)
		return
	}
	token := c.Request.Header.Get("Authorization")
	r, err := CreateTrans(token, cont.Coll, obj)
	if err != nil {
		HandleHttpError(c, err, application.Logger)
		return
	}
	Id := reflect.ValueOf(r).Elem().FieldByName("Id").Int()

	c.Header("Location", strconv.FormatInt(Id, 10))

	c.JSON(http.StatusCreated, Messages["ResourceCreated"])
	c.Abort()

	return
}

func (cont *Controller) Rel(c *gin.Context) {
	var arrIds []int
	c.BindJSON(&arrIds)
	field := c.Param("field")
	origin_id, _ := strconv.Atoi(c.Param("id"))

	token := c.Request.Header.Get("Authorization")

	err := cont.Coll.Rel(token, origin_id, field, arrIds)
	if err != nil {
		HandleHttpError(c, err, application.Logger)
		return
	}
	c.JSON(http.StatusOK, Messages["ResourceUpdated"])
}

func (cont *Controller) UnRel(c *gin.Context) {
	var arrIds []int
	c.BindJSON(&arrIds)

	field := c.Param("field")
	origin_id, _ := strconv.Atoi(c.Param("id"))
	token := c.Request.Header.Get("Authorization")

	err := cont.Coll.UnRel(token, origin_id, field, arrIds)
	if err != nil {
		HandleHttpError(c, err, application.Logger)
		return
	}
	c.JSON(http.StatusOK, Messages["ResourceUpdated"])
}

func (cont *Controller) UpdateRel(c *gin.Context) {
	var arrIds []int
	c.BindJSON(&arrIds)
	field := c.Param("field")
	origin_id, _ := strconv.Atoi(c.Param("id"))

	token := c.Request.Header.Get("Authorization")

	err := cont.Coll.UpdateRel(token, origin_id, field, arrIds)
	if err != nil {
		HandleHttpError(c, err, application.Logger)
		return
	}

	c.JSON(http.StatusOK, Messages["ResourceUpdated"])
}

func (cont *Controller) GetRel(c *gin.Context) {

	field := c.Param("field")
	origin_id, _ := strconv.Atoi(c.Param("id"))
	token := c.Request.Header.Get("Authorization")

	result, err := cont.Coll.GetRel(token, origin_id, field)
	if err != nil {
		HandleHttpError(c, err, application.Logger)
		return
	}
	c.JSON(http.StatusOK, result)
}
func (cont *Controller) GetBase() string {
	return cont.Base
}
func (cont *Controller) SetBase(base string) {
	cont.Base = base
}

/*
* This is a function that receives query string values and create params object for filter
* @param 	interface{}		instance of struct that needs filter
* @param 	url.Values 		query parameters
* @return	Params 			param object
 */
func (cont *Controller) ReadParams(c *gin.Context) *Params {
	vals := c.Request.URL.Query()
	instance := cont.Coll.GetInstance()
	return ReadHttpParams(vals, instance)
}
