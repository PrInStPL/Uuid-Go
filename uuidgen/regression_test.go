package uuidgen

import "testing"

// TestRegressionParse pins what parse accepts and rejects, so refactoring the
// format detection cannot change it.
func TestRegressionParse(t *testing.T) {
	const v4 = "3f2a8b6c-1d4e-4f5a-9b7c-0d1e2f3a4b5c"
	const v7 = "0192a0b1-c2d3-7e4f-8a5b-6c7d8e9fa0b1"
	tests := []struct {
		input string
		want  string // canonical UUID, or "" when parse must fail
	}{
		{v4, v4},
		{"3F2A8B6C-1D4E-4F5A-9B7C-0D1E2F3A4B5C", v4},
		{"{" + v4 + "}", v4},
		{"urn:uuid:" + v4, v4},
		{"3f2a8b6c1d4e4f5a9b7c0d1e2f3a4b5c", v4},
		{"3F2A8B6C1D4E4F5A9B7C0D1E2F3A4B5C", v4},
		{"PyqLbB1OT1qbfA0eLzpLXA==", v4},
		{"PyqLbB1OT1qbfA0eLzpLXA", v4},
		{"83962268023154358113187420214925085532", v4},
		{"00083962268023154358113187420214925085532", v4},
		{v7, v7},
		{"AZKgscLTfk+KW2x9jp+gsQ==", v7},
		{"AZKgscLTfk-KW2x9jp-gsQ", v7},
		{"2090562606348121737443780115287744689", v7},
		// Digit-only input is an integer, never dash-less hex.
		{"11111111111141111111111111111111", "0000008c-3def-b1ef-59da-673864d471c7"},
		{"0", "00000000-0000-0000-0000-000000000000"},
		{"", ""},
		{"not-a-uuid", ""},
		{"-2090562606348121737443780115287744689", ""},
		{"+2090562606348121737443780115287744689", ""},
		{"340282366920938463463374607431768211456", ""},
		{"AZKgscLTfk+KW2x9jp+gsQ", ""},   // standard alphabet without padding
		{"AZKgscLTfk-KW2x9jp-gsQ==", ""}, // URL-safe alphabet with padding
		{"AZKgscLT", ""},
		{"0192a0b1c2d37e4f8a5b6c7d8e9fa0b", ""},
		{"0192a0b1-c2d3-7e4f-8a5b-6c7d8e9fa0bz", ""},
	}
	for _, tt := range tests {
		got, err := parse(tt.input)
		if tt.want == "" {
			if err == nil {
				t.Errorf("parse(%q) = %s, want error", tt.input, got)
			}
			continue
		}
		if err != nil || got.String() != tt.want {
			t.Errorf("parse(%q) = %s, %v; want %s", tt.input, got, err, tt.want)
		}
	}
}
