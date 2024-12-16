// Package auth предоставляет функциональность для работы с аутентификацией пользователей через куки.
package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
)

const (
	// cookieName - название куки для хранения идентификатора пользователя.
	cookieName = "user_id"
	// secretKey - секретный ключ для подписи идентификатора пользователя.
	secretKey = "super-secret-key"
)

// GenerateUserID генерирует новый уникальный идентификатор пользователя.
func GenerateUserID() string {
	return uuid.New().String()
}

// SignUserID создает подпись для идентификатора пользователя.
func SignUserID(userID string) string {
	h := hmac.New(sha256.New, []byte(secretKey))
	h.Write([]byte(userID))
	return hex.EncodeToString(h.Sum(nil))
}

// SetUserCookie устанавливает пользователю подписанную куку с его идентификатором.
func SetUserCookie(w http.ResponseWriter, userID string) {
	signature := SignUserID(userID)
	cookieValue := userID + "|" + signature
	cookie := &http.Cookie{
		Name:     cookieName,
		Value:    cookieValue,
		Expires:  time.Now().Add(24 * time.Hour * 365), // Кука действует 1 год
		HttpOnly: true,
		Secure:   false,
	}
	http.SetCookie(w, cookie)
}

// ValidateUserCookie проверяет подлинность куки и возвращает идентификатор пользователя.
func ValidateUserCookie(r *http.Request) (string, bool) {
	cookie, err := r.Cookie(cookieName)
	if err != nil {
		return "", false
	}

	parts := strings.Split(cookie.Value, "|")
	if len(parts) != 2 {
		return "", false
	}

	userID := parts[0]
	signature := parts[1]

	expectedSignature := SignUserID(userID)
	if hmac.Equal([]byte(expectedSignature), []byte(signature)) {
		return userID, true
	}

	return "", false
}
