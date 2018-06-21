package hermes

import (
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetFilterParams(t *testing.T) {

	v := url.Values{}
	v.Set("name", "mahsa")
	v.Add("family", "ghoreishi")
	v.Add("email", "m.ghoreishi1@gmail.com")
	v.Add("age$from", "23")
	v.Add("age$to", "27")
	v.Add("child_count", "1,2,3")
	v.Add("male", "true")

	v.Add("notexist", "33")

	params := ReadHttpParams(v, &Person{})
	filterParams := params.List

	assert.Equal(t, Filter{Type: "exact", FieldType: "string", Value: "mahsa"}, filterParams["Name"])
	assert.Equal(t, Filter{Type: "exact", FieldType: "string", Value: "ghoreishi"}, filterParams["Family"])
	assert.Equal(t, Filter{Type: "exact", FieldType: "string", Value: "m.ghoreishi1@gmail.com"}, filterParams["Email"])
	assert.Equal(t, Filter{Type: "exact", FieldType: "bool", Value: true}, filterParams["Male"])
	assert.Equal(t, Filter{Type: "range", FieldType: "int", Value: RangeFilter{From: 23, To: 27}}, filterParams["Age"])
	assert.Equal(t, Filter{Type: "array", FieldType: "int", Value: []int{1, 2, 3}}, filterParams["Child_Count"])
	assert.Equal(t, Filter{Type: "", FieldType: "", Value: nil}, filterParams["notexist"])

	v = url.Values{}
	v.Add("age$from", "23")
	params = ReadHttpParams(v, Person{})
	filterParams = params.List

	assert.Equal(t, Filter{Type: "range", FieldType: "int", Value: RangeFilter{From: 23}}, filterParams["Age"])

	v = url.Values{}
	v.Add("age$to", "23")
	v.Add("register_date$from", "2016-02-22T18:24:49.193177+03:30")
	params = ReadHttpParams(v, Person{})
	filterParams = params.List

	assert.Equal(t, Filter{Type: "range", FieldType: "time", Value: RangeFilter{From: "2016-02-22T18:24:49.193177+03:30"}}, filterParams["Register_Date"])
	assert.Equal(t, Filter{Type: "range", FieldType: "int", Value: RangeFilter{To: 23}}, filterParams["Age"])

}
