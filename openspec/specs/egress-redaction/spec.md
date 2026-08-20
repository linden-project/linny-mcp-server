# egress-redaction Specification

## Purpose
TBD - created by archiving change egress-redaction. Update Purpose after archive.
## Requirements
### Requirement: Credentials are redacted from text

The redactor SHALL detect and replace common credential shapes with a typed
`[REDACTED:<kind>]` placeholder, at minimum: PEM private-key blocks, AWS access-key
IDs, AWS secret-access-key assignments, GitHub tokens, JWTs, and generic
`key|secret|token|password` assignments. The raw secret SHALL NOT survive in the
output.

#### Scenario: PEM private key removed

- **WHEN** text contains a `-----BEGIN … PRIVATE KEY-----` … `-----END … PRIVATE
  KEY-----` block
- **THEN** the block is replaced by a placeholder and none of the key material remains

#### Scenario: AWS access key redacted

- **WHEN** text contains an `AKIA…`-shaped access-key ID
- **THEN** it is replaced by `[REDACTED:aws-access-key]`

#### Scenario: Token assignment value redacted, key kept

- **WHEN** text contains `token: ghp_…`
- **THEN** the value is replaced by a placeholder while the `token:` key remains for
  context

### Requirement: Structured responses are deep-scrubbed

The redactor SHALL provide a value walker that scrubs every string within an
arbitrary response value (strings, maps, and slices), so no field of a structured
tool response can carry a secret.

#### Scenario: Nested map and slice scrubbed

- **WHEN** a response value is a map whose nested slice contains a secret-bearing
  string
- **THEN** the walked value has that secret redacted and its structure otherwise
  unchanged

### Requirement: Applied to all tool responses

All MCP tool responses SHALL pass through the redactor before serialization, so no
response can return a credential regardless of the request. (Wiring is consumed by
the read-tool capability.)

#### Scenario: Indexed secret note is scrubbed on read

- **WHEN** a note containing fake credentials is indexed and then read back through
  the redactor
- **THEN** the returned body contains no AWS key, private key, or token material

### Requirement: Reports redaction count without leaking the secret

`Redact` SHALL return the number of redactions performed alongside the redacted text,
and SHALL NOT include the removed secret in that report.

#### Scenario: Count reported

- **WHEN** text with two distinct secrets is redacted
- **THEN** the reported count is at least two and the secrets are absent from the
  output

