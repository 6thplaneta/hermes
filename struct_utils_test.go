package hermes

import (
	"github.com/stretchr/testify/assert"
	"reflect"
	"testing"
	"time"
)

var tagValues = []struct {
	tag, key, returnVal string
}{
	{"dbspace:persons,searchable,editable", "dbspace", "persons"},
	{"dbspace:,searchable,editable", "dbspace", ""},
	{"dbspace:,searchable,editable", "searchable", ""},
	{"dbspace:,searchable,editable", "editable", ""},
	{"dbspace:,searchable,editable,", "editable", ""},
	{"searchable,editable,dbspace:persons", "dbspace", "persons"},
	{"searchable,editable,dbspace:persons,", "dbspace", "persons"},
	{"dbspace:persons", "dbspace", "persons"},
}

func TestGetValueOfTagByKey(t *testing.T) {
	for i := 0; i < len(tagValues); i++ {
		val := GetValueOfTagByKey(tagValues[i].tag, tagValues[i].key)
		assert.Equal(t, val, tagValues[i].returnVal)
	}
}

func TestGetFieldsByTag(t *testing.T) {
	type Person struct {
		Id     int
		Name   string `validate:"required"`
		Family string `validate:"required"`
		Email  string `validate:"type:string"`
		Age    int    `validate:"required"`
	}
	var person Person
	person = Person{}
	//it should work with pointer of object
	arrFields := GetFieldsByTag(&person, "validate", "required")
	assert.Contains(t, arrFields, "Name", "Family", "Age")

	arrFields = GetFieldsByTag(person, "validate", "required")
	assert.NotContains(t, arrFields, "Id", "Age", "Email")

	arrFields = GetFieldsByTag(person, "validate", "type:string")
	assert.Contains(t, arrFields, "Email")

}

func TestGetFeildExistanceAndType(t *testing.T) {
	type User struct {
		Email     string
		LoginDate time.Time `hermes:"type:time"`
	}
	type Person struct {
		Id        int
		Name      string
		Family    string
		Email     string `hermes:"type:time"`
		Age       int
		BirthDate time.Time `hermes:"type:time"`
		User      User
	}
	person := Person{}
	//it should work with address of object
	isexists, typ := GetFieldExistanceAndType(&person, "Name")
	assert.Equal(t, true, isexists)
	assert.Equal(t, "string", typ)

	isexists, typ = GetFieldExistanceAndType(&person, "Email")
	assert.Equal(t, true, isexists)
	assert.NotEqual(t, "time", typ)

	isexists, typ = GetFieldExistanceAndType(&person, "BirthDate")
	assert.Equal(t, true, isexists)
	assert.Equal(t, "time", typ)

	isexists, typ = GetFieldExistanceAndType(&person, "NotExistField")
	assert.Equal(t, false, isexists)
	assert.Equal(t, "", typ)

	//test inner objects
	isexists, typ = GetFieldExistanceAndType(&person, "User.Email")
	assert.Equal(t, true, isexists)
	assert.Equal(t, "string", typ)

	isexists, typ = GetFieldExistanceAndType(&person, "User.LoginDate")
	assert.Equal(t, true, isexists)
	assert.Equal(t, "time", typ)
}

func TestConcatArrayToString(t *testing.T) {
	type Person struct {
		Id   int
		Name string
	}

	arr := []Person{{Id: 1, Name: "mahsa"}, {Id: 2, Name: "elham"}, {Id: 3, Name: "mahshid"}}
	result := ConcatArrayToString(arr, "Id", ",")
	assert.Equal(t, "1,2,3", result)

	result = ConcatArrayToString(arr, "id", "-")
	assert.Equal(t, "1-2-3", result)
}
func TestGetFieldType(t *testing.T) {
	type User struct {
		Email     string
		LoginDate time.Time `hermes:"type:time"`
	}
	type Person struct {
		Id   int
		Name string
		User User
	}

	var person Person
	//it should also work with address of object
	var err error
	rtype, _ := GetFieldType(&person, "User")
	assert.Equal(t, "hermes.User", rtype.String())

	rtype, err = GetFieldType(person, "User.Email")
	assert.NoError(t, err)
	assert.Equal(t, "string", rtype.String())

	rtype, err = GetFieldType(person, "User.LoginDate")
	assert.NoError(t, err)
	assert.Equal(t, "time.Time", rtype.String())

	rtype, err = GetFieldType(person, "User1.Email")
	assert.Error(t, err, "Property does not exists")

	rtype, err = GetFieldType(person, "User.Email1")
	assert.Error(t, err, "Property does not exists")
}

func TestGetTagValue(t *testing.T) {
	type User struct {
		Email      string    `hermes:"type:"`
		Login_Date time.Time `hermes:"type:time"`
	}
	user := User{}
	var val string
	val, _ = GetTagValue(user, "Login_Date", "hermes", "type")
	assert.Equal(t, "time", val)

	val, _ = GetTagValue(user, "email", "hermes", "type")
	assert.Equal(t, "", val)
	val, _ = GetTagValue(user, "notexistfield", "hermes", "type")
	assert.Equal(t, "", val)

}

func TestHasValue(t *testing.T) {
	type User struct {
		Id   int
		Name string
	}
	users := []User{{Name: "Mahsa"}, {Name: "Mahshid"}, {Name: "Sara"}}
	assert.Equal(t, true, HasValue(reflect.ValueOf(users), "Name", "Mahsa"))
	//case sensitive
	assert.Equal(t, false, HasValue(reflect.ValueOf(users), "Name", "mahsa"))

	assert.Equal(t, false, HasValue(reflect.ValueOf(users), "Name", "elham"))
	assert.Equal(t, false, HasValue(reflect.ValueOf(users), "Id", "Mahsa"))

}
