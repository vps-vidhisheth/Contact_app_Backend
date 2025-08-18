package modules

import (
	"Contact_App/component/auth"
	userCtrl "Contact_App/component/user/controller"

	"github.com/gorilla/mux"
)

func RegisterUserRoutes(r *mux.Router) {

	r.HandleFunc("/login", userCtrl.LoginHandler).Methods("POST")
	r.HandleFunc("/user", userCtrl.CreateUserHandler).Methods("POST")

	user := r.PathPrefix("/user").Subrouter()
	user.Use(auth.MiddlewareAdminActive)

	user.HandleFunc("", userCtrl.GetAllUsersHandler).Methods("GET")
	user.HandleFunc("/{userID}", userCtrl.GetUserByIDHandler).Methods("GET")
	user.HandleFunc("/{userID}", userCtrl.UpdateUserHandler).Methods("PUT")
	user.HandleFunc("/{userID}", userCtrl.DeleteUserHandler).Methods("DELETE")
}
