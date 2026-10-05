package uuidgen

import (
	"encoding/base64"
	"errors"
	"fmt"
	"math/big"
	"strings"

	"github.com/google/uuid"
)

// Supported reports whether the UUID version can be generated and validated.
func Supported(version int) bool {
	return version == 4 || version == 7
}

// Generate returns a UUID for the requested version (4 or 7).
func Generate(version int) (uuid.UUID, error) {
	switch version {
	case 4:
		return uuid.NewRandom()
	case 7:
		return uuid.NewV7()
	default:
		return uuid.Nil, errors.New("unsupported UUID version: use 4 or 7")
	}
}

// CheckFormat reports whether the format and letter case options are valid.
func CheckFormat(format string, letterCase string) error {
	switch strings.ToLower(format) {
	case "human":
		switch strings.ToLower(letterCase) {
		case "lower", "upper":
			return nil
		default:
			return errors.New("unsupported case for human format: use lower or upper")
		}
	case "base64", "int":
		return nil
	default:
		return errors.New("unsupported format: use human, base64, or int")
	}
}

// Format renders the UUID using the specified format: "human" (default), "base64", or "int".
// The letterCase parameter affects only the human format ("lower" or "upper").
func Format(id uuid.UUID, format string, letterCase string) (string, error) {
	if err := CheckFormat(format, letterCase); err != nil {
		return "", err
	}
	switch strings.ToLower(format) {
	case "base64":
		return base64.StdEncoding.EncodeToString(id[:]), nil
	case "int":
		return new(big.Int).SetBytes(id[:]).Text(10), nil
	default:
		if strings.ToLower(letterCase) == "upper" {
			return strings.ToUpper(id.String()), nil
		}
		return id.String(), nil
	}
}

// Validate checks whether the input represents an RFC 9562 UUID (human, base64, or int formats).
// If version is non-zero (4 or 7), validation is constrained to that version.
// When version is 0, the function accepts either UUIDv4 or UUIDv7.
func Validate(input string, version int) (uuid.Version, error) {
	if version != 0 && !Supported(version) {
		return 0, errors.New("invalid version constraint: use 4 or 7")
	}

	candidate, err := parse(input)
	if err != nil {
		return 0, err
	}
	if candidate.Variant() != uuid.RFC4122 {
		return 0, fmt.Errorf("invalid UUID: unsupported variant %s", candidate.Variant())
	}

	candidateVersion := candidate.Version()
	if version != 0 {
		if int(candidateVersion) != version {
			return candidateVersion, errors.New("UUID version does not match constraint")
		}
	} else if !Supported(int(candidateVersion)) {
		return candidateVersion, errors.New("unsupported UUID version: expected 4 or 7")
	}

	return candidateVersion, nil
}

// parse detects the input format. Digit-only input is always treated as an integer,
// so a 32-digit number is never misread as dash-less hex.
func parse(input string) (uuid.UUID, error) {
	if isDigits(input) {
		return parseInt(input)
	}
	if parsed, err := uuid.Parse(input); err == nil {
		return parsed, nil
	}
	if decoded, err := base64.StdEncoding.DecodeString(input); err == nil && len(decoded) == 16 {
		return uuid.FromBytes(decoded)
	}
	return uuid.Nil, errors.New("invalid UUID: unsupported format or parse failure")
}

func parseInt(input string) (uuid.UUID, error) {
	num, ok := new(big.Int).SetString(input, 10)
	if !ok || num.BitLen() > 128 {
		return uuid.Nil, errors.New("invalid UUID: integer out of 128-bit range")
	}
	var buf [16]byte
	num.FillBytes(buf[:])
	return uuid.UUID(buf), nil
}

func isDigits(s string) bool {
	if s == "" {
		return false
	}
	for _, r := range s {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}
