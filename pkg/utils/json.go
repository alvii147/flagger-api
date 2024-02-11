package utils

import (
	"fmt"
	"strconv"
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
