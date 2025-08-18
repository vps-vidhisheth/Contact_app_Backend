package auth

import (
	"Contact_App/helper"
	"net/http"
)

func MiddlewareAdminActive(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		claims := GetUserClaims(r)
		if claims == nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		if !helper.IsAuthorizedAdmin(helper.UserData{
			IsAdmin:  claims.IsAdmin,
			IsActive: claims.IsActive,
		}) {
			http.Error(w, "Forbidden: admin privileges required", http.StatusForbidden)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func MiddlewareStaffActive(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		claims := GetUserClaims(r)
		if claims == nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		if !helper.IsAuthorizedStaff(helper.UserData{
			IsAdmin:  claims.IsAdmin,
			IsActive: claims.IsActive,
		}) {
			http.Error(w, "Forbidden: staff privileges required", http.StatusForbidden)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func MiddlewareUserActive(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		claims := GetUserClaims(r)
		if claims == nil || !claims.IsActive {
			http.Error(w, "Unauthorized or inactive user", http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}
