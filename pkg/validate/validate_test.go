package validate_test

import (
	"testing"

	"github.com/alvii147/flagger-api/pkg/validate"
)

type Req struct {
	Email     string `validate:"email"`
	Password  string `validate:"minlength:8"`
	FirstName string `validate:"minlength:1,maxlength:50"`
	LastName  string `validate:"minlength:1,maxlength:50"`
}

func TestValidate(t *testing.T) {
	req := &Req{}
	v := validate.NewValidator()
	v.ValidateStringEmail("email", req.Email)
	v.ValidateStringMinLength("first_name", req.FirstName, 1)
	for _, e := range v.Failures() {
		t.Log(e)
	}

	req = &Req{
		Email:     "name@example.com",
		FirstName: "FirstName",
		LastName:  "LastName",
		Password:  "93n9emDSFDS39",
	}
	v = validate.NewValidator()
	v.ValidateStringEmail("email", req.Email)
	v.ValidateStringMinLength("first_name", req.FirstName, 1)
	for _, e := range v.Failures() {
		t.Log(e)
	}
}
