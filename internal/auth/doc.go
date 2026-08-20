// Package auth defines the Authenticator interface and the v1
// StaticTokenAuthenticator. Tokens are compared with
// crypto/subtle.ConstantTimeCompare; failures yield a bare 401 with no timing
// signal. An OIDCAuthenticator may be added later without changing callers.
package auth
