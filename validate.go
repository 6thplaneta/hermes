package hermes

import (
	"bytes"
	"gopkg.in/bluesuncorp/validator.v6"
	"strings"
)

func ValidateField(value interface{}, rule string) error {
	var validate *validator.Validate

	config := validator.Config{
		TagName:         "validate",
		ValidationFuncs: validator.BakedInValidators,
	}
	validate = validator.New(config)
	err := validate.Field(value, rule)

	return err

}
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
		case "required":
			buff.WriteString(err.Field + " is " + err.Tag + "! ")
		case "email":
			buff.WriteString(err.Field + " should be as " + err.Tag + " format! ")
		default:
			buff.WriteString(err.Field + " is " + err.Tag + "! ")

		}
	}

	erN := ErrObjectInvalid
	erN.Decsription = strings.TrimSpace(buff.String())

	return erN
}
