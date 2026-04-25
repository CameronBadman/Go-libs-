package generaljwt

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

type fakeVerifier struct {
	claims *Claims
	err    error
}

func (f fakeVerifier) Verify(ctx context.Context, token string) (*Claims, error) {
	if f.err != nil {
		return nil, f.err
	}

	return f.claims, nil
}

func TestRequireAuthMissingAuthorizationHeader(t *testing.T) {
	mw := New(fakeVerifier{
		claims: &Claims{Subject: "user-123"},
	})

	handler := mw.RequireAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("next handler should not be called")
	}))

	req := httptest.NewRequest(http.MethodGet, "/private", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected status %d, got %d", http.StatusUnauthorized, rec.Code)
	}
}

func TestRequireAuthInvalidAuthorizationScheme(t *testing.T) {
	mw := New(fakeVerifier{
		claims: &Claims{Subject: "user-123"},
	})

	handler := mw.RequireAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("next handler should not be called")
	}))

	req := httptest.NewRequest(http.MethodGet, "/private", nil)
	req.Header.Set("Authorization", "Basic abc123")

	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected status %d, got %d", http.StatusUnauthorized, rec.Code)
	}
}

func TestRequireAuthEmptyBearerToken(t *testing.T) {
	mw := New(fakeVerifier{
		claims: &Claims{Subject: "user-123"},
	})

	handler := mw.RequireAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("next handler should not be called")
	}))

	req := httptest.NewRequest(http.MethodGet, "/private", nil)
	req.Header.Set("Authorization", "Bearer ")

	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected status %d, got %d", http.StatusUnauthorized, rec.Code)
	}
}

func TestRequireAuthInvalidToken(t *testing.T) {
	mw := New(fakeVerifier{
		err: errors.New("invalid token"),
	})

	handler := mw.RequireAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("next handler should not be called")
	}))

	req := httptest.NewRequest(http.MethodGet, "/private", nil)
	req.Header.Set("Authorization", "Bearer invalid-token")

	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected status %d, got %d", http.StatusUnauthorized, rec.Code)
	}
}

func TestRequireAuthValidTokenCallsNextAndStoresClaims(t *testing.T) {
	expectedClaims := &Claims{
		Subject:  "user-123",
		Username: "alice",
		Groups:   []string{"admin"},
		Scopes:   []string{"read", "write"},
	}

	mw := New(fakeVerifier{
		claims: expectedClaims,
	})

	called := false

	handler := mw.RequireAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true

		claims, ok := ClaimsFromContext(r.Context())
		if !ok {
			t.Fatal("expected claims in context")
		}

		if claims.Subject != expectedClaims.Subject {
			t.Fatalf("expected subject %q, got %q", expectedClaims.Subject, claims.Subject)
		}

		w.WriteHeader(http.StatusNoContent)
	}))

	req := httptest.NewRequest(http.MethodGet, "/private", nil)
	req.Header.Set("Authorization", "Bearer valid-token")

	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if !called {
		t.Fatal("expected next handler to be called")
	}

	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected status %d, got %d", http.StatusNoContent, rec.Code)
	}
}

func TestOptionalAuthNoTokenCallsNextWithoutClaims(t *testing.T) {
	mw := New(fakeVerifier{
		claims: &Claims{Subject: "user-123"},
	})

	handler := mw.OptionalAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, ok := ClaimsFromContext(r.Context()); ok {
			t.Fatal("did not expect claims in context")
		}

		w.WriteHeader(http.StatusNoContent)
	}))

	req := httptest.NewRequest(http.MethodGet, "/public", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected status %d, got %d", http.StatusNoContent, rec.Code)
	}
}

func TestOptionalAuthValidTokenStoresClaims(t *testing.T) {
	mw := New(fakeVerifier{
		claims: &Claims{Subject: "user-123"},
	})

	handler := mw.OptionalAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		claims, ok := ClaimsFromContext(r.Context())
		if !ok {
			t.Fatal("expected claims in context")
		}

		if claims.Subject != "user-123" {
			t.Fatalf("expected subject %q, got %q", "user-123", claims.Subject)
		}

		w.WriteHeader(http.StatusNoContent)
	}))

	req := httptest.NewRequest(http.MethodGet, "/public", nil)
	req.Header.Set("Authorization", "Bearer valid-token")

	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected status %d, got %d", http.StatusNoContent, rec.Code)
	}
}
