# Uuid-Go

Command-line tool for generating and validating UUIDv4 and UUIDv7 (RFC 9562).

## Build

```sh
go build -o uuid-go .
```

Build metadata shown by `--help` can be set with `-ldflags`:

```sh
go build -o uuid-go -ldflags "\
  -X main.buildVersion=1.0.0 \
  -X main.buildDate=$(date -u +%Y-%m-%d) \
  -X main.buildNumber=42" .
```

## Usage

```sh
uuid-go                    # UUIDv4, lowercase (default)
uuid-go -7                 # UUIDv7 (same as --uuid=7)
uuid-go --case=upper       # uppercase
uuid-go --format=base64    # 16 raw bytes, standard base64
uuid-go --format=int       # unsigned 128-bit decimal integer
uuid-go --validate=<id>    # validate a UUID, auto-detecting its format
uuid-go --validate=<id> -7 # validate and require version 7
uuid-go --help             # usage and build information
```

`--validate` accepts the same three formats it can print:

- **human** – `xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx`, any letter case; also the
  dash-less hex, `{...}` and `urn:uuid:...` forms accepted by
  [`github.com/google/uuid`](https://pkg.go.dev/github.com/google/uuid#Parse).
- **base64** – standard base64 of exactly 16 bytes.
- **int** – unsigned decimal; input made only of digits is always treated as an integer.

A UUID is valid when it has the RFC 9562 variant and is version 4 or 7
(or the version given with `-4`, `-7` or `--uuid`).

Errors go to stderr with exit code 1; invalid flags exit with code 2.
