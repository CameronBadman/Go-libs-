// Package generaljwt provides reusable standard-library HTTP JWT middleware.
package generaljwt

import (
	"context"
	"errors"
	"net/http"
	"strings"
)

type Claims struct {
	Subject  string
	Username string
	Groups   []string
	Scopes   []string
	Raw      map[string]any
}

type TokenVerifier interface {
	Verify(ctx context.Context, token string) (*Claims, error)
}

type Middleware struct {
	Verifier TokenVerifier
}

type claimsContextKey struct{}

func New(verifier TokenVerifier) Middleware {
	return Middleware{
		Verifier: verifier,
	}
}

func (m Middleware) RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if m.Verifier == nil {
			http.Error(w, "auth verifier not configured", http.StatusInternalServerError)
			return
		}

		token, err := BearerToken(r)
		if err != nil {
			http.Error(w, "missing bearer token", http.StatusUnauthorized)
			return
		}

		claims, err := m.Verifier.Verify(r.Context(), token)
		if err != nil {
			http.Error(w, "invalid token", http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), claimsContextKey{}, claims)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (m Middleware) OptionalAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if m.Verifier == nil {
			next.ServeHTTP(w, r)
			return
		}

		token, err := BearerToken(r)
		if err != nil {
			next.ServeHTTP(w, r)
			return
		}

		claims, err := m.Verifier.Verify(r.Context(), token)
		if err != nil {
			next.ServeHTTP(w, r)
			return
		}

		ctx := context.WithValue(r.Context(), claimsContextKey{}, claims)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func ClaimsFromContext(ctx context.Context) (*Claims, bool) {
	claims, ok := ctx.Value(claimsContextKey{}).(*Claims)
	return claims, ok
}

func MustClaimsFromContext(ctx context.Context) *Claims {
	claims, ok := ClaimsFromContext(ctx)
	if !ok {
		panic("generaljwt: claims missing from context")
	}

	return claims
}

func BearerToken(r *http.Request) (string, error) {
	header := r.Header.Get("Authorization")
	if header == "" {
		return "", errors.New("missing Authorization header")
	}

	token, ok := strings.CutPrefix(header, "Bearer ")
	if !ok {
		return "", errors.New("Authorization header must use Bearer scheme")
	}

	token = strings.TrimSpace(token)
	if token == "" {
		return "", errors.New("empty bearer token")
	}

	return token, nil
}
