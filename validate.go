package hermes

import (
	"bytes"
	"gopkg.in/bluesuncorp/validator.v6"
	"strings"
)

//checks validation rules on a field
func ValidateField(field interface{}, rule string) error {
	var validate *validator.Validate

	config := validator.Config{
		TagName:         "validate",
		ValidationFuncs: validator.BakedInValidators,
	}
	validate = validator.New(config)
	err := validate.Field(field, rule)

	return err

}

//checks validation rules on all fields of an object
func ValidateStruct(value interface{}) error {

	var validate *validator.Validate

	config := validator.Config{
		TagName:         "validate",
		ValidationFuncs: validator.BakedInValidators,
	}

	validate = validator.New(config)
	errs := validate.Struct(value)
	if len(errs) == 0 {
		return nil
	}

	buff := bytes.NewBufferString("")
	for _, err := range errs {
		switch err.Tag {
		//return appropriate messages
		case "required":
			//err.Feild is name of field
			buff.WriteString(err.Field + " is " + err.Tag + "! ")
		case "email":
			buff.WriteString(err.Field + " should be as " + err.Tag + " format! ")
		default:
			buff.WriteString(err.Field + " is " + err.Tag + "! ")

		}
	}

	// concat all errors in one error object and return it
	erN := ErrObjectInvalid
	erN.Decsription = strings.TrimSpace(buff.String())

	return erN
}
