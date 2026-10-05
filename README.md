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
toolchain are shown instead. A build from a git checkout uses the commit's
VCS data; `go install …@<commit or branch>` has no VCS data, so the commit date
(UTC) and abbreviated revision are read from the module pseudo-version
(e.g. `v0.0.0-20261005090906-4e4593dd695e`). A tagged release such as `v1.2.3`
carries no date, and the build number is only ever set via `-ldflags`.

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
uuid-go <id>                  # validate a UUID, auto-detecting its format
uuid-go --format=hex <id>     # validate and convert to another format
uuid-go --validate=<id>       # validate a UUID, auto-detecting its format
uuid-go --validate=<id> -7    # validate and require version 7
uuid-go --validate=<id> <id>  # validate several values
uuid-go --validate=- < ids    # validate one value per line from stdin
uuid-go --help                # usage and build information
```

`--case` applies to the `human` and `hex` formats.

### Conversion

```sh
uuid-go [--format=<format>] [--case=lower|upper] [-4|-7|--uuid=N] [--pure] <value>
```

The value is processed in three steps, and the first failing step stops with
an error on stderr, exit code 1 and nothing on stdout:

1. its format is recognised by pattern (any input format listed below),
2. it is validated (RFC 9562 variant, version 4 or 7, or the forced version),
3. it is converted to `--format` (with `--case` for `human` and `hex`).

Without `--format` only steps 1–2 run and the validation result is printed.

| | default | `--pure` |
|---|---|---|
| with `--format` | the converted value and a newline | the converted value only, no newline |
| without `--format` | `valid uuid version N` | nothing — only the exit code (errors are silent too) |

```sh
id=$(uuid-go --pure --format=base64url 0192a0b1-c2d3-7e4f-8a5b-6c7d8e9fa0b1)
if uuid-go --pure -7 "$value"; then echo "UUIDv7"; fi
uuid-go --format=human -- -KGyw9TlT2CKe4ydDh8qOw  # "--" before a value starting with "-"
```

Options must come before the value. `-n` and `--validate` cannot be combined
with a value-based conversion; use `--validate` for several values or stdin.

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
