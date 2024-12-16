package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/uuid"
)

func TestGenerateUserID(t *testing.T) {
	userID := GenerateUserID()
	if _, err := uuid.Parse(userID); err != nil {
		t.Errorf("GenerateUserID returned invalid UUID: %v", err)
	}
}

func TestSignUserID(t *testing.T) {
	userID := "test-user-id"
	signature := SignUserID(userID)

	h := hmac.New(sha256.New, []byte("super-secret-key"))
	h.Write([]byte(userID))
	expected := hex.EncodeToString(h.Sum(nil))

	if signature != expected {
		t.Errorf("SignUserID returned unexpected signature: got %s, want %s", signature, expected)
	}
}

func TestSetUserCookie(t *testing.T) {
	userID := "test-user-id"
	w := httptest.NewRecorder()

	SetUserCookie(w, userID)
	resp := w.Result()
	defer resp.Body.Close()
	cookie := resp.Cookies()

	if len(cookie) == 0 {
		t.Fatal("Expected cookie to be set, but none found")
	}

	if cookie[0].Name != "user_id" {
		t.Errorf("Unexpected cookie name: got %s, want %s", cookie[0].Name, "user_id")
	}

	parts := strings.Split(cookie[0].Value, "|")
	if len(parts) != 2 {
		t.Errorf("Cookie value should contain two parts separated by '|', got: %s", cookie[0].Value)
	}

	signature := SignUserID(parts[0])
	if parts[1] != signature {
		t.Errorf("Invalid signature in cookie: got %s, want %s", parts[1], signature)
	}
}

func TestValidateUserCookie_ValidCookie(t *testing.T) {
	userID := "test-user-id"
	w := httptest.NewRecorder()
	SetUserCookie(w, userID)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	resp := w.Result()
	cookie := resp.Cookies()[0]
	defer resp.Body.Close()
	req.AddCookie(cookie)

	validatedUserID, valid := ValidateUserCookie(req)
	if !valid {
		t.Fatal("Expected cookie to be valid")
	}

	if validatedUserID != userID {
		t.Errorf("Unexpected userID: got %s, want %s", validatedUserID, userID)
	}
}

func TestValidateUserCookie_InvalidCookie(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(&http.Cookie{
		Name:  "user_id",
		Value: "invalid|signature",
	})

	_, valid := ValidateUserCookie(req)
	if valid {
		t.Error("Expected cookie to be invalid, but it was valid")
	}
}

func TestValidateUserCookie_MissingCookie(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)

	_, valid := ValidateUserCookie(req)
	if valid {
		t.Error("Expected cookie to be invalid due to missing cookie, but it was valid")
	}
}

func TestValidateUserCookie_InvalidFormat(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(&http.Cookie{
		Name:  "user_id",
		Value: "invalid-format",
	})

	_, valid := ValidateUserCookie(req)
	if valid {
		t.Error("Expected cookie to be invalid due to incorrect format, but it was valid")
	}
}
