# Uuid-Go

Command-line tool for generating and validating UUIDv4 and UUIDv7 (RFC 9562).

## Install

```sh
go install github.com/PrInStPL/Uuid-Go@latest
```

The binary is installed as `Uuid-Go`. To build it locally as `uuid-go` instead:

```sh
go build -o uuid-go .
```

Requires Go 1.21 or newer.

Build metadata shown by `--help` can be set with `-ldflags`:

```sh
go build -o uuid-go -ldflags "\
  -X main.buildVersion=1.0.0 \
  -X main.buildDate=$(date -u +%Y-%m-%d) \
  -X main.buildNumber=42" .
```

Without `-ldflags`, the version, commit date and revision embedded by the Go
toolchain are shown instead (e.g. the module version for `go install`).

## Usage

```sh
uuid-go                       # UUIDv4, lowercase (default)
uuid-go -7                    # UUIDv7 (same as --uuid=7)
uuid-go -7 -n 5               # five UUIDv7, one per line
uuid-go --case=upper          # uppercase
uuid-go --format=hex          # 32 hex characters without dashes
uuid-go --format=base64       # 16 raw bytes, standard base64
uuid-go --format=base64url    # 16 raw bytes, URL-safe base64 without padding
uuid-go --format=int          # unsigned 128-bit decimal integer
uuid-go --validate=<id>       # validate a UUID, auto-detecting its format
uuid-go --validate=<id> -7    # validate and require version 7
uuid-go --validate=<id> <id>  # validate several values
uuid-go --validate=- < ids    # validate one value per line from stdin
uuid-go --help                # usage and build information
```

`--case` applies to the `human` and `hex` formats.

### Validation

`--validate` accepts every format it can print:

- **human** – `xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx`, any letter case; also the
  `{...}` and `urn:uuid:...` forms accepted by
  [`github.com/google/uuid`](https://pkg.go.dev/github.com/google/uuid#Parse).
- **hex** – 32 hex characters without dashes.
- **base64** – standard base64 of exactly 16 bytes (24 characters).
- **base64url** – unpadded URL-safe base64 of exactly 16 bytes (22 characters).
- **int** – unsigned decimal.

Input made only of digits is always treated as an integer, so a hex value
that happens to contain only digits must be written with dashes.

A UUID is valid when it has the RFC 9562 variant and is version 4 or 7
(or the version given with `-4`, `-7` or `--uuid`). For UUIDv7 the embedded
timestamp is printed as well. `--format`, `--case` and `-n` cannot be
combined with `--validate`.

With a single value the result is printed as `valid uuid version N`. With
several values (or `--validate=-`) each one is reported on its own line as
`<value>: valid uuid version N` or `<value>: invalid: <reason>`, and the exit
code is 1 if any value is invalid.

Errors go to stderr with exit code 1; invalid flags exit with code 2.
