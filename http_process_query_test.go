package hermes

import (
	// "github.com/gin-gonic/gin"

	// "fmt"
	// "github.com/stretchr/testify/assert"
	// "net/url"
	"testing"
	"time"
)

func TestGetFilterParams(t *testing.T) {

	type Person struct {
		Id            int       `json:"id" hermes:"dbspace:offers"`
		Name          string    `json:"name"`
		Family        string    `json:"family"`
		Email         string    `json:"email"`
		Register_Date time.Time `hermes:"type:time" json:"register_date"`
		Age           int       `json:"age"`
		Child_Count   int       `json:"child_count"`
		Male          bool      `json:"male"`
	}

	// personColl, e := NewCollection(Person{}, DBTest())
	// assert.NoError(t, e)
	// cont := NewController(personColl, "")

	// c := *gin.Context{}

	//todo test
	// v := url.Values{}
	// v.Set("name", "mahsa")
	// v.Add("family", "ghoreishi")
	// v.Add("email", "m.ghoreishi1@gmail.com")
	// v.Add("age$from", "23")
	// v.Add("age$to", "27")
	// v.Add("child_count", "1,2,3")
	// v.Add("male", "true")

	// v.Add("notexist", "33")

	// params := ReadHttpParams(v, Person{})
	// filterParams := params.List
	// assert.Equal(t, Filter{Type: "exact", FieldType: "string", Value: "mahsa"}, filterParams["name"])
	// assert.Equal(t, Filter{Type: "exact", FieldType: "string", Value: "ghoreishi"}, filterParams["family"])
	// assert.Equal(t, Filter{Type: "exact", FieldType: "string", Value: "m.ghoreishi1@gmail.com"}, filterParams["email"])
	// assert.Equal(t, Filter{Type: "exact", FieldType: "bool", Value: true}, filterParams["male"])
	// assert.Equal(t, Filter{Type: "range", FieldType: "int", Value: RangeFilter{From: 23, To: 27}}, filterParams["age"])
	// assert.Equal(t, Filter{Type: "array", FieldType: "int", Value: []int{1, 2, 3}}, filterParams["child_count"])
	// assert.Equal(t, Filter{Type: "", FieldType: "", Value: nil}, filterParams["notexist"])

	// v = url.Values{}
	// v.Add("age$from", "23")
	// params = ReadHttpParams(v, Person{})
	// filterParams = params.List

	// assert.Equal(t, Filter{Type: "range", FieldType: "int", Value: RangeFilter{From: 23}}, filterParams["Age"])

	// v = url.Values{}
	// v.Add("age$to", "23")
	// v.Add("register_date$from", "2016-02-22T18:24:49.193177+03:30")
	// params = ReadHttpParams(v, Person{})
	// filterParams = params.List

	// assert.Equal(t, Filter{Type: "range", FieldType: "time", Value: RangeFilter{From: "2016-02-22T18:24:49.193177+03:30"}}, filterParams["Register_Date"])
	// assert.Equal(t, Filter{Type: "range", FieldType: "int", Value: RangeFilter{To: 23}}, filterParams["Age"])

}
