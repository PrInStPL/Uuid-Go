package uuidgen

import (
	"encoding/base64"
	"math/big"
	"strings"
	"testing"

	"github.com/google/uuid"
)

func mustGenerate(t *testing.T, version int) uuid.UUID {
	t.Helper()
	id, err := Generate(version)
	if err != nil {
		t.Fatalf("generate v%d: %v", version, err)
	}
	return id
}

func toInt(id uuid.UUID) string {
	return new(big.Int).SetBytes(id[:]).Text(10)
}

func TestGenerate(t *testing.T) {
	for _, version := range []int{4, 7} {
		id := mustGenerate(t, version)
		if int(id.Version()) != version {
			t.Fatalf("expected version %d, got %d", version, id.Version())
		}
		if id.Variant() != uuid.RFC4122 {
			t.Fatalf("expected RFC 4122 variant, got %s", id.Variant())
		}
	}
}

func TestGenerateInvalidVersion(t *testing.T) {
	for _, version := range []int{0, 1, 5, 8} {
		if _, err := Generate(version); err == nil {
			t.Fatalf("expected error for version %d", version)
		}
	}
}

func TestFormat(t *testing.T) {
	id := uuid.MustParse("0192a0b1-c2d3-7e4f-8a5b-6c7d8e9fa0b1")
	tests := []struct {
		format, letterCase, want string
	}{
		{"human", "lower", "0192a0b1-c2d3-7e4f-8a5b-6c7d8e9fa0b1"},
		{"human", "upper", "0192A0B1-C2D3-7E4F-8A5B-6C7D8E9FA0B1"},
		{"HUMAN", "Upper", "0192A0B1-C2D3-7E4F-8A5B-6C7D8E9FA0B1"},
		{"base64", "lower", base64.StdEncoding.EncodeToString(id[:])},
		{"int", "lower", toInt(id)},
	}
	for _, tt := range tests {
		got, err := Format(id, tt.format, tt.letterCase)
		if err != nil {
			t.Fatalf("Format(%q, %q): unexpected error: %v", tt.format, tt.letterCase, err)
		}
		if got != tt.want {
			t.Fatalf("Format(%q, %q) = %q, want %q", tt.format, tt.letterCase, got, tt.want)
		}
	}
}

func TestFormatInvalid(t *testing.T) {
	id := mustGenerate(t, 4)
	if _, err := Format(id, "unknown", "lower"); err == nil {
		t.Fatal("expected error for invalid format")
	}
	if _, err := Format(id, "human", "mixed"); err == nil {
		t.Fatal("expected error for invalid human case")
	}
}

func TestFormatRoundTrip(t *testing.T) {
	for _, version := range []int{4, 7} {
		id := mustGenerate(t, version)
		for _, format := range []string{"human", "base64", "int"} {
			for _, letterCase := range []string{"lower", "upper"} {
				formatted, err := Format(id, format, letterCase)
				if err != nil {
					t.Fatalf("format %s: %v", format, err)
				}
				got, err := Validate(formatted, version)
				if err != nil {
					t.Fatalf("validate %s %q: %v", format, formatted, err)
				}
				if int(got) != version {
					t.Fatalf("validate %s: expected version %d, got %d", format, version, got)
				}
			}
		}
	}
}

func TestValidateIntFormatMismatch(t *testing.T) {
	id := mustGenerate(t, 7)
	if _, err := Validate(toInt(id), 4); err == nil {
		t.Fatal("expected version mismatch error")
	}
}

func TestValidateRejects(t *testing.T) {
	v7 := mustGenerate(t, 7)
	tests := []struct {
		name       string
		input      string
		constraint int
	}{
		{"version 1 auto-detect", "6ba7b810-9dad-11d1-80b4-00c04fd430c8", 0},
		{"version 1 constraint", "6ba7b810-9dad-11d1-80b4-00c04fd430c8", 1},
		// Digit-only input must be read as an integer, not as dash-less hex (which would yield v4).
		{"32-digit integer", "11111111111141111111111111111111", 0},
		{"non-RFC variant", "00000000-0000-4000-0000-000000000000", 0},
		{"negative integer", "-" + toInt(v7), 0},
		{"signed integer", "+" + toInt(v7), 0},
		{"integer above 128 bits", "340282366920938463463374607431768211456", 0},
		{"nil uuid", "0", 0},
		{"short base64", base64.StdEncoding.EncodeToString(v7[:8]), 0},
		{"empty", "", 0},
		{"garbage", "not-a-uuid", 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if version, err := Validate(tt.input, tt.constraint); err == nil {
				t.Fatalf("expected error for %q, got version %d", tt.input, version)
			}
		})
	}
}

func TestValidateAcceptsLeadingZeroInt(t *testing.T) {
	id := mustGenerate(t, 4)
	if _, err := Validate("000"+toInt(id), 4); err != nil {
		t.Fatalf("expected leading zeros to be accepted: %v", err)
	}
}

func TestValidateHumanUppercase(t *testing.T) {
	id := mustGenerate(t, 4)
	if _, err := Validate(strings.ToUpper(id.String()), 0); err != nil {
		t.Fatalf("expected uppercase to be accepted: %v", err)
	}
}
