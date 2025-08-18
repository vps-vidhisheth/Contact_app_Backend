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

	adminEmail := "admin@example.com"
	adminPass := "admin@123"

	var existingUser user.User
	if err := database.Where("email = ?", adminEmail).First(&existingUser).Error; err != nil {

		hashedPass, err := auth.HashPassword(adminPass)
		if err != nil {
			log.Fatalf("Failed to hash password: %v", err)
		}

		firstAdmin := user.User{
			FName:    "Admin",
			LName:    "User",
			Email:    adminEmail,
			Password: hashedPass,
			IsAdmin:  true,
			IsActive: true,
		}

		if err := database.Create(&firstAdmin).Error; err != nil {
			log.Fatalf("Failed to create admin user: %v", err)
		}

		fmt.Println("First admin user created:")
		fmt.Println("Email:", adminEmail)
		fmt.Println("Password:", adminPass)
	} else {

		hashedPass, err := auth.HashPassword(adminPass)
		if err != nil {
			log.Fatalf("Failed to hash password: %v", err)
		}

		existingUser.Password = hashedPass
		existingUser.IsAdmin = true
		existingUser.IsActive = true

		if err := database.Save(&existingUser).Error; err != nil {
			log.Fatalf("Failed to update admin user: %v", err)
		}

		fmt.Println("Existing admin user password reset:")
		fmt.Println("Email:", adminEmail)
		fmt.Println("Password:", adminPass)
	}

	r := mux.NewRouter()

	r.HandleFunc("/api/user", controller.CreateUserHandler).Methods("POST")
	r.HandleFunc("/api/login", controller.LoginHandler).Methods("POST")

	api := r.PathPrefix("/api").Subrouter()
	api.Use(auth.AttachUserToContextIfTokenValid)

	modules.RegisterUserRoutes(api)

	modules.RegisterContactRoutes(api)
	modules.RegisterContactDetailRoutes(api, database)

	corsHandler := handlers.CORS(
		handlers.AllowedOrigins([]string{"*"}),
		handlers.AllowedMethods([]string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}),
		handlers.AllowedHeaders([]string{"Content-Type", "Authorization"}),
	)(r)

	log.Println("Server is running on :8080")
	log.Fatal(http.ListenAndServe(":8080", corsHandler))
}
