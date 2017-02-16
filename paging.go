package hermes

import (
	// "fmt"
	"github.com/gin-gonic/gin"
	"strconv"
)

type Paging struct {
	Num   int
	Size  int
	Sort  string
	Order string
}

func ReadPaging(c *gin.Context) (*Paging, error) {
	var page int
	var err error
	givenpage := c.Query("$page")
	if givenpage == "" {
		page = 1
	} else {
		page, err = strconv.Atoi(givenpage)
		if err != nil {
			baderr := NewError("BadRequest", err.Error())
			return &Paging{}, baderr
		}
		if page < 1 {
			baderr := NewError("BadRequest", "page must be a possitive number")
			return &Paging{}, baderr
		}

	}
	pageSize, _ := strconv.Atoi(c.Query("$page_size"))
	sortBy := EscapeChars(c.Query("$sort_by"))
	sortOrder := EscapeChars(c.Query("$sort_order"))
	if pageSize == 0 {
		pageSize = 10
	}
	return &Paging{page, pageSize, sortBy, sortOrder}, nil
}
