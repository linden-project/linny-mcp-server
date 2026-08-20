package auth

import (
	"bufio"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

// TokenRecord is one entry in the token file. Only the hash is stored — never
// the raw token value.
type TokenRecord struct {
	Name   string   `json:"name"`
	Hash   string   `json:"hash"` // hex-encoded SHA-256 of the raw token
	Scopes []string `json:"scopes"`
}

// StaticTokenAuthenticator authenticates against a fixed set of hashed tokens.
type StaticTokenAuthenticator struct {
	records []TokenRecord
}

// NewStaticTokenAuthenticator builds an authenticator from token records.
func NewStaticTokenAuthenticator(records []TokenRecord) *StaticTokenAuthenticator {
	return &StaticTokenAuthenticator{records: records}
}

// Authenticate hashes the presented token and compares it, in constant time,
// against every stored hash. It never early-exits on a match, so timing does
// not reveal which record (if any) matched.
func (a *StaticTokenAuthenticator) Authenticate(token string) (Identity, error) {
	if token == "" {
		return Identity{}, ErrUnauthorized
	}
	sum := sha256.Sum256([]byte(token))

	matched := -1
	for i := range a.records {
		want, err := hex.DecodeString(a.records[i].Hash)
		if err != nil || len(want) != len(sum) {
			// Keep the loop's timing uniform even for malformed records.
			subtle.ConstantTimeCompare(sum[:], sum[:])
			continue
		}
		if subtle.ConstantTimeCompare(sum[:], want) == 1 {
			matched = i
		}
	}
	if matched < 0 {
		return Identity{}, ErrUnauthorized
	}
	r := a.records[matched]
	return Identity{Name: r.Name, Scopes: append([]string(nil), r.Scopes...)}, nil
}

// HashToken returns the hex-encoded SHA-256 of a raw token.
func HashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

// GenerateToken returns a fresh base64url token carrying at least 32 bytes of
// CSPRNG entropy (no padding).
func GenerateToken() (string, error) {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(buf), nil
}

// LoadTokenFile reads a token file. The format is JSON-lines: one TokenRecord
// object per line; blank lines and lines beginning with '#' are ignored.
func LoadTokenFile(path string) ([]TokenRecord, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer func() { _ = f.Close() }()

	var records []TokenRecord
	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	line := 0
	for sc.Scan() {
		line++
		text := strings.TrimSpace(sc.Text())
		if text == "" || strings.HasPrefix(text, "#") {
			continue
		}
		var rec TokenRecord
		if err := json.Unmarshal([]byte(text), &rec); err != nil {
			return nil, fmt.Errorf("token file %s line %d: %w", path, line, err)
		}
		if rec.Hash == "" {
			return nil, fmt.Errorf("token file %s line %d: record has empty hash", path, line)
		}
		records = append(records, rec)
	}
	if err := sc.Err(); err != nil {
		return nil, err
	}
	if len(records) == 0 {
		return nil, fmt.Errorf("token file %s contains no records", path)
	}
	return records, nil
}
