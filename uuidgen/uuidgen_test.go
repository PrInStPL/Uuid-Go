package uuidgen

import (
	"encoding/base64"
	"math/big"
	"strings"
	"testing"
)

func TestGenerateVersion4(t *testing.T) {
	id, err := Generate("4")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if id.Version() != 4 {
		t.Fatalf("expected version 4, got %d", id.Version())
	}
}

func TestGenerateVersion7(t *testing.T) {
	id, err := Generate("7")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if id.Version() != 7 {
		t.Fatalf("expected version 7, got %d", id.Version())
	}
}

func TestGenerateInvalidVersion(t *testing.T) {
	if _, err := Generate("1"); err == nil {
		t.Fatal("expected error for invalid version")
	}
}

func TestFormatHuman(t *testing.T) {
	id, _ := Generate("4")
	formatted, err := Format(id, "human", "lower")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if formatted != strings.ToLower(id.String()) {
		t.Fatalf("expected %s, got %s", strings.ToLower(id.String()), formatted)
	}
	if strings.ToUpper(formatted) == formatted {
		t.Fatalf("expected lowercase formatting, got %s", formatted)
	}
}

func TestFormatBase64(t *testing.T) {
	id, _ := Generate("4")
	formatted, err := Format(id, "base64", "lower")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	decoded, err := base64.StdEncoding.DecodeString(formatted)
	if err != nil {
		t.Fatalf("invalid base64 output: %v", err)
	}
	if len(decoded) != 16 {
		t.Fatalf("expected 16 bytes, got %d", len(decoded))
	}
}

func TestFormatInt(t *testing.T) {
	id, _ := Generate("4")
	formatted, err := Format(id, "int", "lower")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if strings.IndexFunc(formatted, func(r rune) bool { return r < '0' || r > '9' }) != -1 {
		t.Fatalf("expected digits only, got %s", formatted)
	}
	num := new(big.Int)
	if _, ok := num.SetString(formatted, 10); !ok {
		t.Fatalf("failed to parse int format: %s", formatted)
	}
}

func TestFormatInvalid(t *testing.T) {
	id, _ := Generate("4")
	if _, err := Format(id, "unknown", "lower"); err == nil {
		t.Fatal("expected error for invalid format")
	}
}

func TestFormatHumanUpper(t *testing.T) {
	id, _ := Generate("4")
	formatted, err := Format(id, "human", "upper")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if formatted != strings.ToUpper(id.String()) {
		t.Fatalf("expected uppercase format %s, got %s", strings.ToUpper(id.String()), formatted)
	}
}

func TestFormatHumanInvalidCase(t *testing.T) {
	id, _ := Generate("4")
	if _, err := Format(id, "human", "mixed"); err == nil {
		t.Fatal("expected error for invalid human case")
	}
}

func TestValidateHuman(t *testing.T) {
	id, _ := Generate("4")
	version, err := Validate(id.String(), "")
	if err != nil {
		t.Fatalf("expected validation success, got %v", err)
	}
	if version != 4 {
		t.Fatalf("expected version 4, got %d", version)
	}
}

func TestValidateBase64WithConstraint(t *testing.T) {
	id, _ := Generate("7")
	encoded := base64.StdEncoding.EncodeToString(id[:])
	version, err := Validate(encoded, "7")
	if err != nil {
		t.Fatalf("expected validation success, got %v", err)
	}
	if version != 7 {
		t.Fatalf("expected version 7, got %d", version)
	}
}

func TestValidateIntFormatMismatch(t *testing.T) {
	id, _ := Generate("7")
	num := new(big.Int).SetBytes(id[:]).Text(10)
	if _, err := Validate(num, "4"); err == nil {
		t.Fatal("expected version mismatch error")
	}
}

func TestValidateUnsupportedVersion(t *testing.T) {
	parsed := "6ba7b810-9dad-11d1-80b4-00c04fd430c8" // Version 1
	if _, err := Validate(parsed, ""); err == nil {
		t.Fatal("expected unsupported version error for auto-detect")
	}
}
