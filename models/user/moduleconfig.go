package user

import (
	"errors"
	"fmt"
	"strings"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func hashPassword(password string) (string, error) {
	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(hashed), err
}

func SeedInitialAdmin(db *gorm.DB) (*User, error) {
	const defaultEmail = "admin.user@example.com"
	const defaultPassword = "admin123"

	var existing User
	err := db.Where("email = ?", defaultEmail).First(&existing).Error
	if err == nil {

		return &existing, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	hashedPass, hashErr := hashPassword(defaultPassword)
	if hashErr != nil {
		return nil, fmt.Errorf("failed to hash password: %v", hashErr)
	}

	admin := &User{
		FName:    "Admin",
		LName:    "User",
		Email:    defaultEmail,
		Password: hashedPass,
		IsAdmin:  true,
		IsActive: true,
	}

	if err := db.Create(admin).Error; err != nil {
		return nil, fmt.Errorf("failed to seed initial admin: %v", err)
	}
	return admin, nil
}

func NewInternalUser(fname, lname string, isAdmin bool) (*User, error) {
	email := fmt.Sprintf("%s.%s@example.com", strings.ToLower(fname), strings.ToLower(lname))

	hashedPass, hashErr := hashPassword("admin123")
	if hashErr != nil {
		return nil, fmt.Errorf("failed to hash password: %v", hashErr)
	}

	return &User{
		FName:    fname,
		LName:    lname,
		Email:    email,
		Password: hashedPass,
		IsAdmin:  isAdmin,
		IsActive: true,
	}, nil
}
