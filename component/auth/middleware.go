package auth

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"golang.org/x/crypto/bcrypt"
)

// This is the secret key used to sign and verify JWT tokens.
var jwtKey = []byte("your-secret-key")

type Claims struct {
	UserID   int  `json:"user_id"`
	IsAdmin  bool `json:"is_admin"`
	IsActive bool `json:"is_active"`
	jwt.RegisteredClaims
}

type contextKey string

// his is a key used for storing and retrieving data in Go’s context.Context.
const userContextKey = contextKey("userClaims")

func GenerateToken(userID int, isAdmin bool, isActive bool) (string, error) {
	expirationTime := time.Now().Add(1 * time.Hour)

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

func AttachUserToContextIfTokenValid(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		publicPaths := map[string]bool{
			"/api/login":  true,
			"/api/signup": true,
		}
		if publicPaths[r.URL.Path] {
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
