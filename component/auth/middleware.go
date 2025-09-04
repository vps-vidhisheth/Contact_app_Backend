// package auth

// import (
// 	"context"
// 	"errors"
// 	"fmt"
// 	"net/http"
// 	"os"
// 	"strings"
// 	"time"

// 	"github.com/golang-jwt/jwt/v4"
// 	"golang.org/x/crypto/bcrypt"
// )

// var jwtKey = []byte(os.Getenv("JWT_SECRET"))

// type Claims struct {
// 	UserID   int  `json:"user_id"`
// 	IsAdmin  bool `json:"is_admin"`
// 	IsActive bool `json:"is_active"`
// 	jwt.RegisteredClaims
// }

// type contextKey string

// const userContextKey = contextKey("userClaims")

// func GenerateToken(userID int, isAdmin bool, isActive bool) (string, error) {
// 	expirationTime := time.Now().Add(1 * time.Hour)

// 	claims := &Claims{
// 		UserID:   userID,
// 		IsAdmin:  isAdmin,
// 		IsActive: isActive,
// 		RegisteredClaims: jwt.RegisteredClaims{
// 			ExpiresAt: jwt.NewNumericDate(expirationTime),
// 			IssuedAt:  jwt.NewNumericDate(time.Now()),
// 			Issuer:    "ContactApp",
// 			Subject:   "user_auth",
// 		},
// 	}

// 	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
// 	return token.SignedString(jwtKey)
// }

// func ParseToken(tokenStr string) (*Claims, error) {
// 	claims := &Claims{}
// 	token, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
// 		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
// 			return nil, errors.New("unexpected signing method")
// 		}
// 		return jwtKey, nil
// 	})

// 	if err != nil {
// 		var ve *jwt.ValidationError
// 		if errors.As(err, &ve) && ve.Errors&jwt.ValidationErrorExpired != 0 {
// 			return nil, errors.New("token expired")
// 		}
// 		return nil, fmt.Errorf("token parse error: %w", err)
// 	}

// 	if !token.Valid {
// 		return nil, errors.New("invalid token")
// 	}
// 	return claims, nil
// }

// func AttachUserToContextIfTokenValid(next http.Handler) http.Handler {
// 	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

// 		if strings.HasPrefix(r.URL.Path, "/api/v1/login") || strings.HasPrefix(r.URL.Path, "/api/v1/signup") {
// 			next.ServeHTTP(w, r)
// 			return
// 		}

// 		authHeader := r.Header.Get("Authorization")
// 		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
// 			http.Error(w, "Unauthorized: missing or invalid Authorization header", http.StatusUnauthorized)
// 			return
// 		}

// 		tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
// 		claims, err := ParseToken(tokenStr)
// 		if err != nil {
// 			http.Error(w, fmt.Sprintf("Unauthorized: %s", err.Error()), http.StatusUnauthorized)
// 			return
// 		}

// 		ctx := context.WithValue(r.Context(), userContextKey, claims)
// 		next.ServeHTTP(w, r.WithContext(ctx))
// 	})
// }

// func GetUserClaims(r *http.Request) *Claims {
// 	claims, ok := r.Context().Value(userContextKey).(*Claims)
// 	if !ok {
// 		return nil
// 	}
// 	return claims
// }

// func HashPassword(password string) (string, error) {
// 	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
// 	return string(bytes), err
// }

// func CheckPasswordHash(password, hash string) bool {
// 	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
// 	return err == nil
// }

// func GetUserContextKey() interface{} {
// 	return userContextKey
// }

package auth

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"golang.org/x/crypto/bcrypt"
)

var jwtKey = []byte(getSecret())

func getSecret() string {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		// fallback to default for local/dev
		secret = "your-secret-key"
	}
	return secret
}

type Claims struct {
	UserID   int  `json:"user_id"`
	IsAdmin  bool `json:"is_admin"`
	IsActive bool `json:"is_active"`
	jwt.RegisteredClaims
}

type contextKey string

const userContextKey = contextKey("userClaims")

// GenerateToken - creates a new signed JWT
func GenerateToken(userID int, isAdmin bool, isActive bool) (string, error) {
	expirationTime := time.Now().Add(24 * time.Hour)

	claims := &Claims{
		UserID:   userID,
		IsAdmin:  isAdmin,
		IsActive: isActive,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "ContactApp",
			Subject:   "user_auth",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtKey)
}

// ParseToken - validates and parses JWT
func ParseToken(tokenStr string) (*Claims, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return jwtKey, nil
	})

	if err != nil {
		var ve *jwt.ValidationError
		if errors.As(err, &ve) && ve.Errors&jwt.ValidationErrorExpired != 0 {
			return nil, errors.New("token expired")
		}
		return nil, fmt.Errorf("token parse error: %w", err)
	}

	if !token.Valid {
		return nil, errors.New("invalid token")
	}
	return claims, nil
}

// Middleware - attach user to context if token is valid
func AttachUserToContextIfTokenValid(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/api/v1/login") || strings.HasPrefix(r.URL.Path, "/api/v1/signup") {
			next.ServeHTTP(w, r)
			return
		}

		authHeader := r.Header.Get("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			http.Error(w, "Unauthorized: missing or invalid Authorization header", http.StatusUnauthorized)
			return
		}

		tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
		claims, err := ParseToken(tokenStr)
		if err != nil {
			http.Error(w, fmt.Sprintf("Unauthorized: %s", err.Error()), http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), userContextKey, claims)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func GetUserClaims(r *http.Request) *Claims {
	claims, ok := r.Context().Value(userContextKey).(*Claims)
	if !ok {
		return nil
	}
	return claims
}

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

func GetUserContextKey() interface{} {
	return userContextKey
}
