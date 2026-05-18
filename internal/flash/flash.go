package flash

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
)

type contextKey struct{}

func GetMessage(ctx context.Context) string {
	v, _ := ctx.Value(contextKey{}).(string)
	return v
}

func Set(w http.ResponseWriter, message string) {
	encrypted, err := encrypt(message)
	if err != nil {
		return
	}
	http.SetCookie(w, &http.Cookie{
		Name:     "flash",
		Value:    encrypted,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
		MaxAge:   60,
	})
}

func Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("flash")
		if err == nil {
			http.SetCookie(w, &http.Cookie{
				Name:   "flash",
				Value:  "",
				Path:   "/",
				MaxAge: -1,
			})
			if message, err := decrypt(cookie.Value); err == nil {
				r = r.WithContext(context.WithValue(r.Context(), contextKey{}, message))
			}
		}
		next.ServeHTTP(w, r)
	})
}

func encrypt(plaintext string) (string, error) {
	key, err := getKey()
	if err != nil {
		return "", err
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}
	ciphertext := gcm.Seal(nonce, nonce, []byte(plaintext), nil)
	return base64.URLEncoding.EncodeToString(ciphertext), nil
}

func decrypt(encoded string) (string, error) {
	key, err := getKey()
	if err != nil {
		return "", err
	}
	data, err := base64.URLEncoding.DecodeString(encoded)
	if err != nil {
		return "", err
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return "", fmt.Errorf("ciphertext too short")
	}
	nonce, ciphertext := data[:nonceSize], data[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", err
	}
	return string(plaintext), nil
}

func getKey() ([]byte, error) {
	return hex.DecodeString(os.Getenv("FLASH_SECRET"))
}
