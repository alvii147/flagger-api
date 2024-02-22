package validate

import (
	"fmt"
	"net/mail"
	"regexp"
	"strings"
	"unicode/utf8"
)

// reSlug is a compiled regular expression for slug string validation.
var reSlug = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)

// Validator validates given values and accumulates validation errors.
type Validator struct {
	failures map[string][]string
}

// NewValidator creates and returns a new Validator.
func NewValidator() *Validator {
	return &Validator{
		failures: make(map[string][]string),
	}
}

// addFailure records a validation failure.
func (v *Validator) addFailure(field string, format string, args ...any) {
	v.failures[field] = append(v.failures[field], fmt.Sprintf(format, args...))
}

// Failures returns recorded validation failures.
func (v *Validator) Failures() map[string][]string {
	return v.failures
}

// Passed returns whether or not all validations have passed.
func (v *Validator) Passed() bool {
	return len(v.failures) == 0
}

// ValidateStringNotBlank validates that a given string is not blank.
func (v *Validator) ValidateStringNotBlank(field string, value string) {
	if utf8.RuneCountInString(strings.TrimSpace(value)) < 1 {
		v.addFailure(field, "\"%s\" cannot be blank", field)
	}
}

// ValidateStringMaxLength validates that a given string is at most of a given length.
func (v *Validator) ValidateStringMaxLength(field string, value string, maxLen int) {
	if utf8.RuneCountInString(value) > maxLen {
		v.addFailure(field, "\"%s\" cannot be more than %d characters long", field, maxLen)
	}
}

// ValidateStringMinLength validates that a given string is at least of a given length.
func (v *Validator) ValidateStringMinLength(field string, value string, minLen int) {
	if utf8.RuneCountInString(value) < minLen {
		v.addFailure(field, "\"%s\" must be at least %d characters long", field, minLen)
	}
}

// ValidateStringEmail validates the format of a given email address.
func (v *Validator) ValidateStringEmail(field string, email string) {
	_, err := mail.ParseAddress(email)
	if err != nil {
		v.addFailure(field, "\"%s\" must be a valid email address", field)
	}
}

// ValidateStringSlug validates that a given string is a valid slug.
func (v *Validator) ValidateStringSlug(field string, value string) {
	if !reSlug.MatchString(value) {
		v.addFailure(field, "\"%s\" must be a slug", field)
	}
}
