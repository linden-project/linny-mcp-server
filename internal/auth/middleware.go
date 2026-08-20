package auth

import "net/http"

// Middleware wraps next with bearer-token authentication. On any failure it
// responds 401 with an empty body and no distinguishing detail. On success it
// injects the resolved Identity into the request context.
func Middleware(a Authenticator, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token, ok := ParseBearer(r.Header.Get("Authorization"))
		if !ok {
			unauthorized(w)
			return
		}
		id, err := a.Authenticate(token)
		if err != nil {
			unauthorized(w)
			return
		}
		next.ServeHTTP(w, r.WithContext(WithIdentity(r.Context(), id)))
	})
}

// unauthorized writes a bare 401 with no body — no hint about the failure mode.
func unauthorized(w http.ResponseWriter) {
	w.WriteHeader(http.StatusUnauthorized)
}
