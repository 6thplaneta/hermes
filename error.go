package hermes

import (
	"database/sql"
	// "errors"
	"fmt"
)

type Error struct {
	Key         string
	Decsription string
}

func (e Error) Error() string {
	return fmt.Sprintf("%s: %s", e.Key, e.Decsription)
}

func NewError(key, desc string) Error {
	return Error{Key: key, Decsription: desc}
}

func newError(key, desc string) Error {

	return Error{Key: key, Decsription: desc}
}

func UserError(desc string) Error {
	return Error{Key: "BadRequest", Decsription: desc}
}

var ErrObjectInvalid = newError("NotValid", "Invalid input")
var ErrDuplicate = newError("DuplicateData", "Duplicate data!")
var ErrDuplicateIndex = newError("DuplicateIndex", "pq: duplicate key value violates unique constraint")
var ErrNotFound = newError("NotFound", "Resource not found!")
var ErrPassFormat = newError("PasswordFormatFailed", "Password format failed!")
var ErrPassword = newError("PasswordError", "Password is not correct!")
var ErrPassRequired = newError("RequiredPassword", "Password is required!")
var ErrAgentNotActive = newError("AgentIsNotActive", "Agent is not activated!")
var ErrTokenInvalid = newError("InvalidToken", "Invalid token!")
var ErrBadRequest = newError("BadRequest", "Bad request")
var ErrInternalServer = newError("InternalServerError", "Internal server error occured!")
var ErrRateExceed = newError("RateExceed", "Rate of request exceeds server policy")
var ErrManyItems = newError("ManyItems", "Too many items!")
var ErrForbidden = newError("Forbidden", "You are not authorized for access to this resource")
var ErrNoDBSpace = newError("NoDbSpace", "dbspace not found in id!")
var ErrJsonFormat = newError("JsonFormatError", "Incorrect Json Format!")
var ErrFieldNotExists = newError("FieldNotExists", "mentioned field does not exists!")

var ErrNoRows = sql.ErrNoRows
