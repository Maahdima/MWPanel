package schema

import (
	"encoding/json"
	"testing"
)

func TestResellerSettingsJSONIntegerCompatibility(t *testing.T) {
	for _, tc := range []struct {
		name, value string
		want        *int64
		invalid     bool
	}{
		{name: "missing", value: `{}`},
		{name: "null", value: `{"rate_rial":null}`},
		{name: "integer", value: `{"rate_rial":9007199254740993}`, want: integerPointer(9007199254740993)},
		{name: "string", value: `{"rate_rial":"9223372036854775807"}`, want: integerPointer(9223372036854775807)},
		{name: "fraction", value: `{"rate_rial":1.5}`, invalid: true},
		{name: "overflow", value: `{"rate_rial":"9223372036854775808"}`, invalid: true},
		{name: "boolean", value: `{"rate_rial":true}`, invalid: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var input ResellerSettings
			err := json.Unmarshal([]byte(tc.value), &input)
			if tc.invalid {
				if err == nil {
					t.Fatal("invalid integer accepted")
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if (input.RateRial == nil) != (tc.want == nil) || input.RateRial != nil && *input.RateRial != *tc.want {
				t.Fatal("integer precision or optional value changed")
			}
		})
	}
}
func integerPointer(v int64) *int64 { return &v }
