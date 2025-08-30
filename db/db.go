package db

import (
	"fmt"
	"log"
	"os"

	"Contact_App/models/contact"
	"Contact_App/models/contact_detail"
	"Contact_App/models/user"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var DB *gorm.DB

func InitDB() {
	dsn := getDSN()

	var err error
	DB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	err = DB.AutoMigrate(
		&user.User{},
		&contact.Contact{},
		&contact_detail.ContactDetail{},
	)
	if err != nil {
		log.Fatalf("AutoMigrate failed: %v", err)
	}

	log.Println("Database connected and models migrated successfully.")
}

func GetDB() *gorm.DB {
	if DB == nil {
		log.Fatal("DB not initialized. Call InitDB() first.")
	}
	return DB
}

// Data Source Name
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

	return fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true&charset=utf8mb4", dbUser, dbPass, dbHost, dbPort, dbName)
}

func SeedInitialAdmin() (*user.User, error) {
	var admin user.User
	if err := DB.Where("email = ?", "admin@example.com").First(&admin).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("admin123"), bcrypt.DefaultCost)
			admin = user.User{
				FName:    "Admin",
				LName:    "User",
				Email:    "admin@example.com",
				Password: string(hashedPassword),
				IsAdmin:  true,
				IsActive: true,
			}
			if err := DB.Create(&admin).Error; err != nil {
				return nil, err
			}
			log.Println("Initial admin user created successfully.")
			return &admin, nil
		}
		return nil, err
	}
	log.Println("Admin user already exists.")
	return &admin, nil
}
