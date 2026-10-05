package uuidgen

import (
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"regexp"
	"strings"
	"time"

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
	case "human", "hex":
		switch strings.ToLower(letterCase) {
		case "lower", "upper":
			return nil
		default:
			return errors.New("unsupported case: use lower or upper")
		}
	case "base64", "base64url", "int":
		return nil
	default:
		return errors.New("unsupported format: use human, hex, base64, base64url, or int")
	}
}

// Format renders the UUID using the specified format: "human" (default), "hex",
// "base64", "base64url", or "int".
// The letterCase parameter affects only the human and hex formats ("lower" or "upper").
func Format(id uuid.UUID, format string, letterCase string) (string, error) {
	if err := CheckFormat(format, letterCase); err != nil {
		return "", err
	}
	var value string
	switch strings.ToLower(format) {
	case "base64":
		return base64.StdEncoding.EncodeToString(id[:]), nil
	case "base64url":
		return base64.RawURLEncoding.EncodeToString(id[:]), nil
	case "int":
		return new(big.Int).SetBytes(id[:]).Text(10), nil
	case "hex":
		value = hex.EncodeToString(id[:])
	default:
		value = id.String()
	}
	if strings.ToLower(letterCase) == "upper" {
		return strings.ToUpper(value), nil
	}
	return value, nil
}

// Validate checks whether the input represents an RFC 9562 UUID in any format
// produced by Format, and returns the parsed UUID.
// If version is non-zero (4 or 7), validation is constrained to that version.
// When version is 0, the function accepts either UUIDv4 or UUIDv7.
func Validate(input string, version int) (uuid.UUID, error) {
	if version != 0 && !Supported(version) {
		return uuid.Nil, errors.New("invalid version constraint: use 4 or 7")
	}

	candidate, err := parse(input)
	if err != nil {
		return uuid.Nil, err
	}
	if candidate.Variant() != uuid.RFC4122 {
		return uuid.Nil, fmt.Errorf("invalid UUID: unsupported variant %s", candidate.Variant())
	}

	candidateVersion := int(candidate.Version())
	if version != 0 {
		if candidateVersion != version {
			return uuid.Nil, fmt.Errorf("UUID version %d does not match constraint %d", candidateVersion, version)
		}
	} else if !Supported(candidateVersion) {
		return uuid.Nil, fmt.Errorf("unsupported UUID version %d: expected 4 or 7", candidateVersion)
	}

	return candidate, nil
}

// Timestamp returns the creation time encoded in a UUIDv7 (the first 48 bits, in Unix milliseconds).
func Timestamp(id uuid.UUID) (time.Time, bool) {
	if id.Version() != 7 {
		return time.Time{}, false
	}
	var ms int64
	for _, b := range id[:6] {
		ms = ms<<8 | int64(b)
	}
	return time.UnixMilli(ms).UTC(), true
}

var errUnrecognized = errors.New("invalid UUID: unsupported format or parse failure")

// Input patterns, checked in order by Detect. Digit-only input is matched first,
// so a 32-digit number is always an integer, never dash-less hex.
var patterns = []struct {
	format string
	re     *regexp.Regexp
}{
	{"int", regexp.MustCompile(`^[0-9]+$`)},
	{"human", regexp.MustCompile(`^(?:(?i:urn:uuid:)` + humanPattern + `|\{` + humanPattern + `\}|` + humanPattern + `)$`)},
	{"hex", regexp.MustCompile(`^[0-9a-fA-F]{32}$`)},
	{"base64", regexp.MustCompile(`^[A-Za-z0-9+/]{22}==$`)},
	{"base64url", regexp.MustCompile(`^[A-Za-z0-9_-]{22}$`)},
}

const humanPattern = `[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}`

// Detect returns the format of the input ("human", "hex", "base64", "base64url" or "int")
// by matching it against each format's pattern. It does not check that the value is a valid UUID.
func Detect(input string) (string, error) {
	for _, p := range patterns {
		if p.re.MatchString(input) {
			return p.format, nil
		}
	}
	return "", errUnrecognized
}

// Convert validates the input like Validate and renders it in the given format.
func Convert(input string, version int, format, letterCase string) (string, error) {
	if err := CheckFormat(format, letterCase); err != nil {
		return "", err
	}
	id, err := Validate(input, version)
	if err != nil {
		return "", err
	}
	return Format(id, format, letterCase)
}

// parse decodes the input according to the format reported by Detect.
func parse(input string) (uuid.UUID, error) {
	format, err := Detect(input)
	if err != nil {
		return uuid.Nil, err
	}
	switch format {
	case "int":
		return parseInt(input)
	case "base64", "base64url":
		enc := base64.StdEncoding
		if format == "base64url" {
			enc = base64.RawURLEncoding
		}
		decoded, err := enc.DecodeString(input)
		if err != nil {
			return uuid.Nil, errUnrecognized
		}
		return uuid.FromBytes(decoded)
	default:
		parsed, err := uuid.Parse(input)
		if err != nil {
			return uuid.Nil, errUnrecognized
		}
		return parsed, nil
	}
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
