package main

import (
	"Contact_App/component/auth"
	"Contact_App/component/user/controller"
	"Contact_App/db"
	"Contact_App/models/user"
	"Contact_App/modules"
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
)

func main() {

	db.InitDB()
	database := db.GetDB()

	admin, err := user.SeedInitialAdmin(database)
	if err != nil {
		log.Fatalf("Error seeding admin user: %v", err)
	}

	fmt.Println("Admin user ready:")
	fmt.Println("Email:", admin.Email)

	r := mux.NewRouter()

	// Public routes with /api/v1 prefix
	r.HandleFunc("/api/v1/user", controller.CreateUserHandler).Methods("POST")
	r.HandleFunc("/api/v1/login", controller.LoginHandler).Methods("POST")

	// Protected routes under /api/v1
	api := r.PathPrefix("/api/v1").Subrouter()
	api.Use(auth.AttachUserToContextIfTokenValid)

	// Register module routes
	modules.RegisterUserRoutes(api)
	modules.RegisterContactRoutes(api)
	modules.RegisterContactDetailRoutes(api, database)

	// CORS setup
	corsHandler := handlers.CORS(
		handlers.AllowedOrigins([]string{"*"}),
		handlers.AllowedMethods([]string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}),
		handlers.AllowedHeaders([]string{"Content-Type", "Authorization"}),
	)(r)

	log.Println("Server is running on :8080")
	log.Fatal(http.ListenAndServe(":8080", corsHandler))
}
