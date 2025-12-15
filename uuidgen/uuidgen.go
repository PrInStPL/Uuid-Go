package uuidgen

import (
	"encoding/base64"
	"errors"
	"math/big"
	"strconv"
	"strings"

	"github.com/google/uuid"
)

// Generate returns a UUID for the requested version ("4" or "7").
func Generate(version string) (uuid.UUID, error) {
	switch version {
	case "4":
		return uuid.NewRandom()
	case "7":
		return uuid.NewV7()
	default:
		return uuid.Nil, errors.New("unsupported UUID version: use 4 or 7")
	}
}

// Format renders the UUID using the specified format: "human" (default), "base64", or "int".
// The letterCase parameter affects only the human format ("lower" or "upper").
func Format(id uuid.UUID, format string, letterCase string) (string, error) {
	switch strings.ToLower(format) {
	case "human":
		value := id.String()
		switch strings.ToLower(letterCase) {
		case "lower":
			return strings.ToLower(value), nil
		case "upper":
			return strings.ToUpper(value), nil
		default:
			return "", errors.New("unsupported case for human format: use lower or upper")
		}
	case "base64":
		return base64.StdEncoding.EncodeToString(id[:]), nil
	case "int":
		num := new(big.Int).SetBytes(id[:])
		return num.Text(10), nil
	default:
		return "", errors.New("unsupported format: use human, base64, or int")
	}
}

// Validate checks whether the input represents a UUID (human, base64, or int formats).
// If version is provided ("4" or "7"), validation is constrained to that version.
// When version is empty, the function accepts either UUIDv4 or UUIDv7.
func Validate(input string, version string) (uuid.Version, error) {
	var candidate uuid.UUID

	if parsed, err := uuid.Parse(input); err == nil {
		candidate = parsed
		goto checkVersion
	}

	if decoded, err := base64.StdEncoding.DecodeString(input); err == nil {
		if parsed, err := uuid.FromBytes(decoded); err == nil {
			candidate = parsed
			goto checkVersion
		}
	}

	if num, ok := new(big.Int).SetString(input, 10); ok {
		bytes := num.Bytes()
		if len(bytes) <= len(candidate) {
			var buf [16]byte
			copy(buf[16-len(bytes):], bytes)
			parsed, err := uuid.FromBytes(buf[:])
			if err == nil {
				candidate = parsed
				goto checkVersion
			}
		}
	}

	return 0, errors.New("invalid UUID: unsupported format or parse failure")

checkVersion:
	candidateVersion := candidate.Version()
	if version != "" {
		expected, err := strconv.Atoi(version)
		if err != nil {
			return 0, errors.New("invalid version constraint: use 4 or 7")
		}
		if int(candidateVersion) != expected {
			return candidateVersion, errors.New("UUID version does not match constraint")
		}
	} else if candidateVersion != 4 && candidateVersion != 7 {
		return candidateVersion, errors.New("unsupported UUID version: expected 4 or 7")
	}

	return candidateVersion, nil
}
