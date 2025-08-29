package main

import (
	"Contact_App/app"
	"Contact_App/db"
	"fmt"
	"log"
)

func main() {
	db.InitDB()
	database := db.GetDB()

	admin, err := db.SeedInitialAdmin()
	if err != nil {
		log.Fatalf("Failed to seed admin user: %v", err)
	}
	fmt.Println("Admin user ready:", admin.Email)

	application := app.NewApp(database)

	application.InitServer(":8080")

	application.StartServer()
}
