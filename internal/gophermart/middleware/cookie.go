package middleware

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"go-musthave-diploma-tpl/internal/gophermart/config"
	"go-musthave-diploma-tpl/internal/gophermart/service"
)

type contextKey string

const UserIDKey contextKey = "userID"

var encryptionKey = []byte(config.EncryptionKey)

func CookieMiddleware(repo *service.GofemartService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Получаем куки
			cookie, err := r.Cookie("userID")
			if err != nil {
				http.Error(w, "Authentication required", http.StatusUnauthorized)
				return
			}

			// Проверяем срок жизни куки
			if !cookie.Expires.IsZero() && cookie.Expires.Before(time.Now()) {
				http.Error(w, "Cookie expired", http.StatusUnauthorized)
				return
			}

			// Пытаемся расшифровать куки
			userIDStr, err := decrypt(cookie.Value)
			if err != nil {
				http.Error(w, "Invalid authentication cookie", http.StatusUnauthorized)
				return
			}

			// Конвертируем строку в число (ID пользователя)
			userID, err := strconv.Atoi(userIDStr)
			if err != nil {
				http.Error(w, "Invalid user ID in cookie", http.StatusUnauthorized)
				return
			}

			// Проверяем что пользователь существует в БД
			user, err := repo.GetUserByID(userID)
			if err != nil || user == nil {
				http.Error(w, "User not found", http.StatusUnauthorized)
				return
			}

			// Всё ок - передаем userID в контекст
			ctx := context.WithValue(r.Context(), UserIDKey, userIDStr)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// SetEncryptedCookie - публичная функция для установки куки из хендлеров
// Используется только при успешной регистрации/логине
func SetEncryptedCookie(w http.ResponseWriter, userID string) {
	encrypted, err := Encrypt(userID)
	if err != nil {
		encrypted = userID
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "userID",
		Value:    encrypted,
		Path:     "/",
		Expires:  time.Now().Add(8 * time.Hour),
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteLaxMode,
	})
}

func Encrypt(userID string) (string, error) {
	block, err := aes.NewCipher(encryptionKey)
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return "", err
	}

	cipherText := gcm.Seal(nonce, nonce, []byte(userID), nil)
	return base64.URLEncoding.EncodeToString(cipherText), nil
}

func decrypt(val string) (string, error) {
	data, err := base64.URLEncoding.DecodeString(val)
	if err != nil {
		return "", err
	}
	block, err := aes.NewCipher(encryptionKey)
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return "", fmt.Errorf("ciphertext to short")
	}
	nonce, cipherText := data[:nonceSize], data[nonceSize:]
	plainText, err := gcm.Open(nil, nonce, cipherText, nil)
	if err != nil {
		return "", err
	}
	return string(plainText), err
}

func GetUserID(ctx context.Context) (string, error) {
	userID, ok := ctx.Value(UserIDKey).(string)
	if !ok {
		return "", fmt.Errorf("user ID not found in context")
	}
	return userID, nil
}
