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
type emailErrorKey struct{}
type emailValueKey struct{}

func GetMessage(ctx context.Context) string {
	v, _ := ctx.Value(contextKey{}).(string)
	return v
}

func GetEmailError(ctx context.Context) string {
	v, _ := ctx.Value(emailErrorKey{}).(string)
	return v
}

func WithEmailError(ctx context.Context, message string) context.Context {
	return context.WithValue(ctx, emailErrorKey{}, message)
}

func GetEmailValue(ctx context.Context) string {
	v, _ := ctx.Value(emailValueKey{}).(string)
	return v
}

func SetEmailValue(w http.ResponseWriter, value string) {
	encrypted, err := encrypt(value)
	if err != nil {
		return
	}
	http.SetCookie(w, &http.Cookie{
		Name:     "flash_email_value",
		Value:    encrypted,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   60,
	})
}

func SetEmailError(w http.ResponseWriter, message string) {
	encrypted, err := encrypt(message)
	if err != nil {
		return
	}
	http.SetCookie(w, &http.Cookie{
		Name:     "flash_email",
		Value:    encrypted,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   60,
	})
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
		SameSite: http.SameSiteLaxMode,
		MaxAge:   60,
	})
}

func Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		if cookie, err := r.Cookie("flash"); err == nil {
			http.SetCookie(w, &http.Cookie{Name: "flash", Value: "", Path: "/", MaxAge: -1})
			if message, err := decrypt(cookie.Value); err == nil {
				ctx = context.WithValue(ctx, contextKey{}, message)
			}
		}
		if cookie, err := r.Cookie("flash_email"); err == nil {
			http.SetCookie(w, &http.Cookie{Name: "flash_email", Value: "", Path: "/", MaxAge: -1})
			if message, err := decrypt(cookie.Value); err == nil {
				ctx = context.WithValue(ctx, emailErrorKey{}, message)
			}
		}
		if cookie, err := r.Cookie("flash_email_value"); err == nil {
			http.SetCookie(w, &http.Cookie{Name: "flash_email_value", Value: "", Path: "/", MaxAge: -1})
			if value, err := decrypt(cookie.Value); err == nil {
				ctx = context.WithValue(ctx, emailValueKey{}, value)
			}
		}
		next.ServeHTTP(w, r.WithContext(ctx))
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
