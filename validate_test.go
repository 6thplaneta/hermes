package hermes

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestValidateField(t *testing.T) {
	type Person struct {
		Id     int
		Name   string
		Family string
		Email  string
		Age    int
	}
	var person Person
	person = Person{Id: 1, Name: "mahsa", Family: "ghoreishi"}

	err := ValidateField(person.Name, "required")
	assert.Error(t, err, "")

	err = ValidateField(person.Age, "required")
	assert.Error(t, err)

	person.Age = 27
	err = ValidateField(person.Age, "required")
	assert.Error(t, err, "")

	person.Email = "m.ghoreishi1@gmail.com"
	err = ValidateField(person.Email, "email")
	assert.Error(t, err, "")

	person.Email = "m.ghoreishi1"
	err = ValidateField(person.Email, "email")
	assert.Error(t, err)
}

func TestValidateStruct(t *testing.T) {
	type Person struct {
		Id     int    `validate:"required"`
		Name   string `validate:"required"`
		Family string `validate:"required"`
		Email  string `validate:"required,email"`
		Age    int    `validate:"required"`
	}
	var person Person
	person = Person{Id: 1, Name: "mahsa", Family: "ghoreishi", Age: 27}
	errs := ValidateStruct(person)
	assert.Equal(t, errs.Error(), "NotValid: Email is required!")

	//incorrect email format
	person.Email = "m.ghoreishi1"
	errs = ValidateStruct(person)
	assert.Equal(t, errs.Error(), "NotValid: Email should be as email format!")

	person.Email = "m.ghoreishi1@gmail.com"
	person.Age = 0
	errs = ValidateStruct(person)
	assert.Equal(t, errs.Error(), "NotValid: Age is required!")

}
