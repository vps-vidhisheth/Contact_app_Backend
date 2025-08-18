package service

import (
	"Contact_App/apperror"
	"Contact_App/models/user"
	"Contact_App/repository"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

var jwtSecret = []byte("your-secret-key")

type CustomClaims struct {
	UserID   int  `json:"user_id"`
	IsAdmin  bool `json:"is_admin"`
	IsActive bool `json:"is_active"`
	jwt.RegisteredClaims
}

func GenerateJWT(u *user.User) (string, error) {
	claims := CustomClaims{
		UserID:   int(u.UserID),
		IsAdmin:  u.IsAdmin,
		IsActive: u.IsActive,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			Issuer:    "contact-app",
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}

func Authenticate(db *gorm.DB, email, password string) (*user.User, error) {
	repo := repository.NewGormRepository()
	uow := repository.NewUnitOfWork(db, true)
	defer uow.Rollback()

	email = strings.ToLower(email)

	var users []user.User
	err := repo.GetAll(uow, &users, repository.Filter("email = ?", email))
	if err != nil || len(users) == 0 {
		return nil, apperror.NewUnauthorized("invalid email or password")
	}

	u := users[0]
	if err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password)); err != nil {
		return nil, apperror.NewUnauthorized("invalid email or password")
	}

	return &u, nil
}

func CreateUser(repo repository.Repository, uow *repository.UnitOfWork, fname, lname string, isAdmin, isActive bool, email, password string) (*user.User, error) {
	if fname == "" || lname == "" || email == "" || password == "" {
		return nil, apperror.NewValidationError("user", "all fields are required")
	}

	hashedPass, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, apperror.NewInternalError("failed to hash password")
	}

	newUser := &user.User{
		FName:    strings.TrimSpace(fname),
		LName:    strings.TrimSpace(lname),
		IsAdmin:  isAdmin,
		IsActive: isActive,
		Email:    strings.ToLower(email),
		Password: string(hashedPass),
	}

	if err := repo.Add(uow, newUser); err != nil {
		return nil, err
	}

	return newUser, nil
}

func GetUserByID(repo repository.Repository, uow *repository.UnitOfWork, userID int) (*user.User, error) {
	var u user.User
	if err := repo.GetByID(uow, uint(userID), &u); err != nil {
		return nil, err
	}
	return &u, nil
}

func GetAllUsers(repo repository.Repository, uow *repository.UnitOfWork) ([]*user.User, error) {
	var users []*user.User
	if err := repo.GetAll(uow, &users); err != nil {
		return nil, err
	}
	return users, nil
}
func UpdateUserByID(repo repository.Repository, uow *repository.UnitOfWork, admin *user.User, userID int, updates *struct {
	FName    string `json:"f_name"`
	LName    string `json:"l_name"`
	IsAdmin  bool   `json:"is_admin"`
	Email    string `json:"email"`
	Password string `json:"password"`
}) (*user.User, error) {
	if !admin.IsAdminActive() {
		return nil, apperror.NewAuthError("update user")
	}

	var existing user.User
	if err := repo.GetByID(uow, uint(userID), &existing); err != nil {
		return nil, err
	}

	if updates.FName != "" {
		existing.FName = updates.FName
	}
	if updates.LName != "" {
		existing.LName = updates.LName
	}
	if updates.Email != "" {
		existing.Email = strings.ToLower(updates.Email)
	}
	if updates.Password != "" {
		hashedPass, _ := bcrypt.GenerateFromPassword([]byte(updates.Password), bcrypt.DefaultCost)
		existing.Password = string(hashedPass)
	}
	existing.IsAdmin = updates.IsAdmin

	if err := repo.Update(uow, &existing); err != nil {
		return nil, err
	}

	return &existing, nil
}

func DeleteUserByID(repo repository.Repository, uow *repository.UnitOfWork, adminID int, userID int) error {
	var admin user.User
	if err := repo.GetByID(uow, uint(adminID), &admin); err != nil {
		return err
	}
	if !admin.IsAdminActive() {
		return apperror.NewAuthError("delete user")
	}
	if adminID == userID {
		return apperror.NewValidationError("user", "admin cannot delete their own account")
	}

	var target user.User
	if err := repo.GetByID(uow, uint(userID), &target); err != nil {
		return err
	}

	if err := uow.DB.Delete(&target).Error; err != nil {
		return apperror.NewInternalError("failed to permanently delete user")
	}

	return nil
}

func ExposeNewUserInternal(repo repository.Repository, uow *repository.UnitOfWork, fname, lname string, isAdmin bool) (*user.User, error) {
	fname = strings.TrimSpace(fname)
	lname = strings.TrimSpace(lname)

	if fname == "" || lname == "" {
		return nil, apperror.NewValidationError("name", "first or last name cannot be empty")
	}

	defaultEmail := fmt.Sprintf("%s.%s@example.com", strings.ToLower(fname), strings.ToLower(lname))
	defaultPassword := "admin123"

	hashedPass, err := bcrypt.GenerateFromPassword([]byte(defaultPassword), bcrypt.DefaultCost)
	if err != nil {
		return nil, apperror.NewInternalError("failed to hash default password")
	}

	u := &user.User{
		FName:    fname,
		LName:    lname,
		IsAdmin:  isAdmin,
		IsActive: true,
		Email:    defaultEmail,
		Password: string(hashedPass),
	}

	if err := repo.Add(uow, u); err != nil {
		return nil, err
	}

	return u, nil
}

func CreateInitialAdminUser(repo repository.Repository, uow *repository.UnitOfWork) (*user.User, error) {
	const defaultEmail = "admin.user@example.com"

	var existing []user.User
	err := repo.GetAll(uow, &existing, repository.Filter("email = ?", defaultEmail))
	if err == nil && len(existing) > 0 {
		return &existing[0], errors.New("admin user already exists")
	}

	return ExposeNewUserInternal(repo, uow, "Admin", "User", true)
}
