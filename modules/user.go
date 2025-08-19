package modules

import (
	"Contact_App/component/auth"
	userCtrl "Contact_App/component/user/controller"

	"github.com/gorilla/mux"
)

func RegisterUserRoutes(r *mux.Router) {
	// Public routes
	r.HandleFunc("/login", userCtrl.LoginHandler).Methods("POST")
	r.HandleFunc("/user", userCtrl.CreateUserHandler).Methods("POST")

	// Admin-only routes
	admin := r.PathPrefix("/user").Subrouter()
	admin.Use(auth.MiddlewareAdminActive)

	admin.HandleFunc("", userCtrl.GetAllUsersHandler).Methods("GET")
	admin.HandleFunc("/{userID}", userCtrl.UpdateUserHandler).Methods("PUT")
	admin.HandleFunc("/{userID}", userCtrl.DeleteUserHandler).Methods("DELETE")

	// Routes for both staff and admin (for example: viewing own user)
	secured := r.PathPrefix("/user").Subrouter()
	secured.Use(auth.MiddlewareAdminActive) // just check JWT, role not required

	secured.HandleFunc("/{userID}", userCtrl.GetUserByIDHandler).Methods("GET")
}
