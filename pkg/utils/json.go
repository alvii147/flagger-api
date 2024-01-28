package utils

import (
	"fmt"
	"net/mail"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// JSONTimeStamp represents time that is convert to unix timestamp in JSON.
type JSONTimeStamp time.Time

// MarshalJSON converts time into unix timestamp bytes.
func (jt JSONTimeStamp) MarshalJSON() ([]byte, error) {
	return []byte(strconv.FormatInt(time.Time(jt).UTC().Unix(), 10)), nil
}

// UnmarshalJSON converts unix timestamp bytes into timestamp.
func (jt *JSONTimeStamp) UnmarshalJSON(p []byte) error {
	s := string(p)
	ts, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return fmt.Errorf("UnmarshalJSON failed to strconv.ParseInt timestamp \"%s\": %w", s, err)
	}

	*(*time.Time)(jt) = time.Unix(ts, 0).UTC()

	return nil
}

// JSONEmail represents email string in JSON that is validated.
type JSONEmail string

// MarshalJSON converts email string into email bytes.
func (je JSONEmail) MarshalJSON() ([]byte, error) {
	return []byte(je), nil
}

// UnmarshalJSON converts email bytes into email string and validates it.
func (je *JSONEmail) UnmarshalJSON(p []byte) error {
	s := string(p)
	unquoted, err := strconv.Unquote(s)
	if err != nil {
		return fmt.Errorf("UnmarshalJSON failed to strconv.Unquote email \"%s\": %w", s, err)
	}

	_, err = mail.ParseAddress(unquoted)
	if err != nil {
		return fmt.Errorf("UnmarshalJSON failed to mail.ParseAddress email \"%s\": %w", unquoted, err)
	}

	*je = JSONEmail(unquoted)

	return nil
}

// JSONNonEmptyString represents non-empty string in JSON.
type JSONNonEmptyString string

// MarshalJSON converts string into bytes.
func (js JSONNonEmptyString) MarshalJSON() ([]byte, error) {
	return []byte(js), nil
}

// UnmarshalJSON converts string bytes into a string and validates its length.
func (js *JSONNonEmptyString) UnmarshalJSON(p []byte) error {
	s := string(p)
	unquoted, err := strconv.Unquote(string(p))
	if err != nil {
		return fmt.Errorf("UnmarshalJSON failed to strconv.Unquote \"%s\": %w", s, err)
	}

	if len(strings.TrimSpace(unquoted)) == 0 {
		return fmt.Errorf("UnmarshalJSON failed, empty string \"%s\": %w", unquoted, err)
	}

	*js = JSONNonEmptyString(unquoted)

	return nil
}

// JSONSlugString represents non-empty string in JSON
// that only allows lowercase alphabets, numbers, and hyphens.
type JSONSlugString string

// MarshalJSON converts string into bytes.
func (js JSONSlugString) MarshalJSON() ([]byte, error) {
	return []byte(js), nil
}

// UnmarshalJSON converts string bytes into a string and validates its length.
func (js *JSONSlugString) UnmarshalJSON(p []byte) error {
	s := string(p)
	unquoted, err := strconv.Unquote(string(p))
	if err != nil {
		return fmt.Errorf("UnmarshalJSON failed to strconv.Unquote email \"%s\": %w", s, err)
	}

	if len(strings.TrimSpace(unquoted)) == 0 {
		return fmt.Errorf("UnmarshalJSON failed, empty string \"%s\": %w", unquoted, err)
	}

	r := regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)
	if !r.MatchString(unquoted) {
		return fmt.Errorf("UnmarshalJSON failed, string \"%s\" is not valid slug: %w", unquoted, err)
	}

	*js = JSONSlugString(unquoted)

	return nil
}
