package db

import (
	"fmt"
	"log"
	"os"

	"Contact_App/models/contact"
	"Contact_App/models/contact_detail"
	"Contact_App/models/user"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var dbInstance *gorm.DB

func InitDB() {
	dsn := getDSN()

	var err error
	dbInstance, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf(" Failed to connect to database: %v", err)
	}

	err = dbInstance.AutoMigrate(
		&user.User{},
		&contact.Contact{},
		&contact_detail.ContactDetail{},
	)
	if err != nil {
		log.Fatalf(" AutoMigrate failed: %v", err)
	}

	log.Println(" Database connected and models migrated successfully.")
}

func GetDB() *gorm.DB {
	if dbInstance == nil {
		log.Fatal(" Attempted to access DB before initialization. Call InitDB() first.")
	}
	return dbInstance
}

func getDSN() string {
	dbUser := os.Getenv("DB_USER")
	dbPass := os.Getenv("DB_PASS")
	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbName := os.Getenv("DB_NAME")

	if dbUser == "" {
		dbUser = "root"
	}
	if dbPass == "" {
		dbPass = "pass@123"
	}
	if dbHost == "" {
		dbHost = "localhost"
	}
	if dbPort == "" {
		dbPort = "3306"
	}
	if dbName == "" {
		dbName = "contact_app_project_structure"
	}

	return fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true", dbUser, dbPass, dbHost, dbPort, dbName)
}
