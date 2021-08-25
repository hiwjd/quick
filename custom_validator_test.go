package quick

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// NewUserRequest is for test
type NewUserRequest struct {
	Username string `validate:"min=3,max=40,regexp=^[a-zA-Z]*$"`
	Name     string `validate:"nonzero"`
	Age      int    `validate:"min=18"`
	Password string `validate:"min=8"`
}

func TestCustomValidator(t *testing.T) {
	v := NewCustomValidator()

	nur := NewUserRequest{Username: "something", Name: "guys", Age: 15, Password: "1234567"}
	errs := v.Validate(nur)
	assert.NotNil(t, errs)

	nur.Age = 18
	nur.Password = "12345678"
	errs = v.Validate(nur)
	assert.Nil(t, errs)
}
