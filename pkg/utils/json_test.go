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

func TestJSONEmailMarshalJSON(t *testing.T) {
	t.Parallel()

	type jsonStruct struct {
		Email utils.JSONEmail `json:"email"`
	}

	p, err := json.Marshal(jsonStruct{
		Email: utils.JSONEmail(`"name@example.com"`),
	})
	require.NoError(t, err)
	require.Regexp(t, `^\s*{\s*"email"\s*:\s*"name@example.com"\s*}\s*$`, string(p))
}

func TestJSONEmailUnmarshalJSONSuccess(t *testing.T) {
	t.Parallel()

	type jsonStruct struct {
		Email utils.JSONEmail `json:"email"`
	}

	s := jsonStruct{}
	err := json.Unmarshal([]byte(`{"email":"name@example.com"}`), &s)
	require.NoError(t, err)
	require.Equal(t, "name@example.com", string(s.Email))
}

func TestJSONEmailUnmarshalJSONError(t *testing.T) {
	t.Parallel()

	type jsonStruct struct {
		Email utils.JSONEmail `json:"email"`
	}

	testcases := []struct {
		name string
		data string
	}{
		{
			name: "Non-string email",
			data: `{"email":314}`,
		},
		{
			name: "Invalid email",
			data: `{"email":"1nv4l1d3m41l"}`,
		},
	}

	for _, testcase := range testcases {
		testcase := testcase
		t.Run(testcase.name, func(t *testing.T) {
			t.Parallel()

			s := jsonStruct{}
			err := json.Unmarshal([]byte(testcase.data), &s)
			require.Error(t, err)
		})
	}
}

func TestJSONNonEmptyStringMarshalJSON(t *testing.T) {
	t.Parallel()

	type jsonStruct struct {
		String utils.JSONNonEmptyString `json:"string"`
	}

	p, err := json.Marshal(jsonStruct{
		String: utils.JSONNonEmptyString(`"deadbeef"`),
	})
	require.NoError(t, err)
	require.Regexp(t, `^\s*{\s*"string"\s*:\s*"deadbeef"\s*}\s*$`, string(p))
}

func TestJSONNonEmptyStringUnmarshalJSONSuccess(t *testing.T) {
	t.Parallel()

	type jsonStruct struct {
		String utils.JSONNonEmptyString `json:"string"`
	}

	s := jsonStruct{}
	err := json.Unmarshal([]byte(`{"string":"deadbeef"}`), &s)
	require.NoError(t, err)
	require.Equal(t, "deadbeef", string(s.String))
}

func TestJSONNonEmptyStringlUnmarshalJSONError(t *testing.T) {
	t.Parallel()

	type jsonStruct struct {
		String utils.JSONNonEmptyString `json:"string"`
	}

	testcases := []struct {
		name string
		data string
	}{
		{
			name: "Empty string",
			data: `{"string":""}`,
		},
		{
			name: "Blank string",
			data: `{"string":"   "}`,
		},
	}

	for _, testcase := range testcases {
		testcase := testcase
		t.Run(testcase.name, func(t *testing.T) {
			t.Parallel()

			s := jsonStruct{}
			err := json.Unmarshal([]byte(testcase.data), &s)
			require.Error(t, err)
		})
	}
}

func TestJSONSlugStringMarshalJSON(t *testing.T) {
	t.Parallel()

	type jsonStruct struct {
		String utils.JSONSlugString `json:"string"`
	}

	p, err := json.Marshal(jsonStruct{
		String: utils.JSONSlugString(`"d34d-b33f"`),
	})
	require.NoError(t, err)
	require.Regexp(t, `^\s*{\s*"string"\s*:\s*"d34d-b33f"\s*}\s*$`, string(p))
}

func TestJSONSlugStringUnmarshalJSONSuccess(t *testing.T) {
	t.Parallel()

	type jsonStruct struct {
		String utils.JSONSlugString `json:"string"`
	}

	s := jsonStruct{}
	err := json.Unmarshal([]byte(`{"string":"d34d-b33f"}`), &s)
	require.NoError(t, err)
	require.Equal(t, "d34d-b33f", string(s.String))
}

func TestJSONSlugStringlUnmarshalJSONError(t *testing.T) {
	t.Parallel()

	type jsonStruct struct {
		String utils.JSONSlugString `json:"string"`
	}

	testcases := []struct {
		name string
		data string
	}{
		{
			name: "Empty string",
			data: `{"string":""}`,
		},
		{
			name: "Blank string",
			data: `{"string":"   "}`,
		},
		{
			name: "String with invalid characters",
			data: `{"string":"hello w*rld"}`,
		},
		{
			name: "String beginning with hyphen",
			data: `{"string":"-d34d-b33f"}`,
		},
	}

	for _, testcase := range testcases {
		testcase := testcase
		t.Run(testcase.name, func(t *testing.T) {
			t.Parallel()

			s := jsonStruct{}
			err := json.Unmarshal([]byte(testcase.data), &s)
			require.Error(t, err)
		})
	}
}
