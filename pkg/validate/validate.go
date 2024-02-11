package validate

import (
	"fmt"
	"net/mail"
	"regexp"
	"strings"
	"unicode/utf8"
)

var reSlug = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)

type Validator struct {
	failures map[string][]string
}

func (v *Validator) failure(field string, msg string) {
	v.failures[field] = append(v.failures[field], msg)
}

func (v *Validator) failuref(field string, format string, args ...any) {
	v.failures[field] = append(v.failures[field], fmt.Sprintf(format, args...))
}

func (v *Validator) Failures() map[string][]string {
	return v.failures
}

func (v *Validator) Passed() bool {
	return len(v.failures) == 0
}

func (v *Validator) ValidatorStringNotBlank(field string, value string) {
	if utf8.RuneCountInString(strings.TrimSpace(value)) < 1 {
		v.failuref(field, "\"%s\" cannot be blank", field)
	}
}

func (v *Validator) ValidateStringMaxLength(field string, value string, maxLen int) {
	if utf8.RuneCountInString(value) > maxLen {
		v.failuref(field, "\"%s\" cannot be more than %d characters long", field, maxLen)
	}
}

func (v *Validator) ValidateStringMinLength(field string, value string, minLen int) {
	if utf8.RuneCountInString(value) < minLen {
		v.failuref(field, "\"%s\" must be at least %d characters long", field, minLen)
	}
}

func (v *Validator) ValidateStringEmail(field string, email string) {
	_, err := mail.ParseAddress(email)
	if err != nil {
		v.failuref(field, "\"%s\" must be a valid email address", field)
	}
}

func (v *Validator) ValidateStringSlug(field string, value string) {
	if !reSlug.MatchString(value) {
		v.failuref(field, "\"%s\" must be a slug", field)
	}
}

func NewValidator() *Validator {
	return &Validator{
		failures: make(map[string][]string),
	}
}
