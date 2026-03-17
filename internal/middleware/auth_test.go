package middleware

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/R4x4nJ031/project-man/internal/config"
	appcontext "github.com/R4x4nJ031/project-man/internal/context"
	jwt "github.com/golang-jwt/jwt/v5"
)

func TestAuthMiddleware_MissingAuthorizationHeader(t *testing.T) {
	cfg := testConfig()
	handler := AuthMiddleware(cfg)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("next handler should not be called")
	}))

	req := httptest.NewRequest(http.MethodGet, "/projects/list", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected status %d, got %d", http.StatusUnauthorized, rec.Code)
	}

	var body map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}

	if body["error"] != "missing authorization header" {
		t.Fatalf("expected missing auth error, got %q", body["error"])
	}
}

func TestAuthMiddleware_ValidTokenInjectsSecurityContext(t *testing.T) {
	cfg := testConfig()
	token := testJWT(t, cfg, time.Now().Add(time.Hour), cfg.JWTIssuer, cfg.JWTAud)

	var gotUserID string
	var gotEmail string

	handler := AuthMiddleware(cfg)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		secCtx := appcontext.GetSecurityContext(r.Context())
		if secCtx == nil {
			t.Fatal("expected security context")
		}

		gotUserID = secCtx.UserID
		gotEmail = secCtx.Email
		w.WriteHeader(http.StatusNoContent)
	}))

	req := httptest.NewRequest(http.MethodGet, "/projects/list", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected status %d, got %d", http.StatusNoContent, rec.Code)
	}

	if gotUserID != "user-123" {
		t.Fatalf("expected user id user-123, got %q", gotUserID)
	}

	if gotEmail != "user@example.com" {
		t.Fatalf("expected email user@example.com, got %q", gotEmail)
	}
}

func TestAuthMiddleware_InvalidAudienceRejected(t *testing.T) {
	cfg := testConfig()
	token := testJWT(t, cfg, time.Now().Add(time.Hour), cfg.JWTIssuer, "wrong-audience")

	handler := AuthMiddleware(cfg)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("next handler should not be called")
	}))

	req := httptest.NewRequest(http.MethodGet, "/projects/list", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected status %d, got %d", http.StatusUnauthorized, rec.Code)
	}

	var body map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}

	if body["error"] != "invalid audience" {
		t.Fatalf("expected invalid audience error, got %q", body["error"])
	}
}

func testConfig() *config.Config {
	return &config.Config{
		Port:      "8080",
		JWTSecret: "test-secret",
		JWTIssuer: "project-man",
		JWTAud:    "project-man-users",
	}
}

func testJWT(t *testing.T, cfg *config.Config, exp time.Time, issuer string, audience string) string {
	t.Helper()

	claims := &CustomClaims{
		UserID: "user-123",
		Email:  "user@example.com",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(exp),
			Issuer:    issuer,
			Audience:  jwt.ClaimStrings{audience},
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(cfg.JWTSecret))
	if err != nil {
		t.Fatalf("sign token: %v", err)
	}

	return signed
}
