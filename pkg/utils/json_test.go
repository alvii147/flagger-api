package utils_test

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/alvii147/flagger-api/pkg/utils"
	"github.com/stretchr/testify/require"
)

func TestJSONTimeStampMarshalJSON(t *testing.T) {
	t.Parallel()

	type jsonStruct struct {
		Timestamp utils.JSONTimeStamp `json:"timestamp"`
	}

	p, err := json.Marshal(jsonStruct{
		Timestamp: utils.JSONTimeStamp(time.Date(2010, 12, 16, 19, 12, 36, 0, time.UTC)),
	})
	require.NoError(t, err)
	require.Regexp(t, `^\s*{\s*"timestamp"\s*:\s*1292526756\s*}\s*$`, string(p))
}

func TestJSONTimeStampUnmarshalJSONSuccess(t *testing.T) {
	t.Parallel()

	type jsonStruct struct {
		Timestamp utils.JSONTimeStamp `json:"timestamp"`
	}

	s := jsonStruct{}
	err := json.Unmarshal([]byte(`{"timestamp":1292526756}`), &s)
	require.NoError(t, err)
	require.Equal(t, time.Date(2010, 12, 16, 19, 12, 36, 0, time.UTC), time.Time(s.Timestamp))
}

func TestJSONTimeStampUnmarshalJSONError(t *testing.T) {
	t.Parallel()

	type jsonStruct struct {
		Timestamp utils.JSONTimeStamp `json:"timestamp"`
	}

	s := jsonStruct{}
	err := json.Unmarshal([]byte(`{"timestamp":"string value"}`), &s)
	require.Error(t, err)
}
