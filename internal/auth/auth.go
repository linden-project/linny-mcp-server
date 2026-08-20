package auth

import (
	"context"
	"errors"
	"strings"
)

// ErrUnauthorized is returned for every authentication failure. It is
// deliberately singular: callers must not distinguish missing, malformed,
// unknown, or wrong tokens.
var ErrUnauthorized = errors.New("unauthorized")

// Identity is the resolved principal behind a valid token.
type Identity struct {
	Name   string
	Scopes []string
}

// Authenticator validates a raw bearer token. StaticTokenAuthenticator is the
// only implementation for v1; an OIDCAuthenticator can be added later without
// changing callers.
type Authenticator interface {
	Authenticate(token string) (Identity, error)
}

// ParseBearer extracts the token from an Authorization header value. It fails
// closed: a missing or malformed header yields ("", false). The scheme match is
// case-insensitive per RFC 7235.
func ParseBearer(header string) (token string, ok bool) {
	const scheme = "bearer "
	if len(header) <= len(scheme) {
		return "", false
	}
	if !strings.EqualFold(header[:len(scheme)], scheme) {
		return "", false
	}
	tok := strings.TrimSpace(header[len(scheme):])
	if tok == "" {
		return "", false
	}
	return tok, true
}

// context plumbing for the resolved identity.
type ctxKey struct{}

// WithIdentity returns a context carrying id.
func WithIdentity(ctx context.Context, id Identity) context.Context {
	return context.WithValue(ctx, ctxKey{}, id)
}

// FromContext returns the identity previously stored with WithIdentity.
func FromContext(ctx context.Context) (Identity, bool) {
	id, ok := ctx.Value(ctxKey{}).(Identity)
	return id, ok
}
