package service

import (
	"Contact_App/apperror"
	"Contact_App/component/auth"
	"Contact_App/helper"
	"Contact_App/models/user"
	"Contact_App/repository"
	"Contact_App/web"
	"errors"
	"fmt"
	"net/http"
	"strconv"
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

func GetUserByID(repo repository.Repository, uow *repository.UnitOfWork, userID int, filters ...repository.QueryProcessor) (*user.User, error) {
	baseFilters := []repository.QueryProcessor{
		repository.Filter("user_id = ?", userID),
	}
	baseFilters = append(baseFilters, filters...)

	var users []user.User
	err := repo.GetAll(uow, &users, baseFilters...)
	if err != nil {
		return nil, err
	}
	if len(users) == 0 {
		return nil, fmt.Errorf("user with ID %d not found", userID)
	}

	return &users[0], nil
}

func GetAllUsers(repo repository.Repository, uow *repository.UnitOfWork, filters ...repository.QueryProcessor) ([]user.User, error) {
	var users []user.User
	err := repo.GetAll(uow, &users, filters...)
	if err != nil {
		return nil, err
	}
	return users, nil
}

func UpdateUserByID(
	repo repository.Repository,
	uow *repository.UnitOfWork,
	claims *auth.Claims,
	userID int,
	updates *struct {
		FName    string `json:"f_name"`
		LName    string `json:"l_name"`
		IsAdmin  bool   `json:"is_admin"`
		Email    string `json:"email"`
		Password string `json:"password"`
	},
) (*user.User, error) {

	var existing user.User
	if err := repo.GetByID(uow, uint(userID), &existing); err != nil {
		return nil, err
	}

	if !claims.IsAdmin {
		updates.IsAdmin = existing.IsAdmin
	}

	updateMap := make(map[string]interface{})
	if updates.FName != "" {
		updateMap["f_name"] = updates.FName
	}
	if updates.LName != "" {
		updateMap["l_name"] = updates.LName
	}
	if updates.Email != "" {
		updateMap["email"] = strings.ToLower(updates.Email)
	}
	if updates.Password != "" {
		hashedPass, _ := bcrypt.GenerateFromPassword([]byte(updates.Password), bcrypt.DefaultCost)
		updateMap["password"] = string(hashedPass)
	}
	updateMap["is_admin"] = updates.IsAdmin

	if err := repo.UpdateWithMap(uow, &user.User{}, updateMap,
		repository.Filter("user_id = ?", userID),
	); err != nil {
		return nil, err
	}

	for k, v := range updateMap {
		switch k {
		case "f_name":
			existing.FName = v.(string)
		case "l_name":
			existing.LName = v.(string)
		case "email":
			existing.Email = v.(string)
		case "password":
			existing.Password = v.(string)
		case "is_admin":
			existing.IsAdmin = v.(bool)
		}
	}

	return &existing, nil
}

// Hard delete (permanently removes user)
func DeleteUserByID(repo repository.Repository, uow *repository.UnitOfWork, adminID int, userID int, hardDelete bool) error {
	var admin user.User
	if err := repo.GetByID(uow, uint(adminID), &admin); err != nil {
		return err
	}

	adminData := helper.UserData{
		IsAdmin:  admin.IsAdmin,
		IsActive: admin.IsActive,
	}

	if !helper.IsAuthorizedAdmin(adminData) {
		return apperror.NewAuthError("delete user")
	}

	if adminID == userID {
		return apperror.NewValidationError("user", "admin cannot delete their own account")
	}

	var target user.User
	if err := repo.GetByID(uow, uint(userID), &target); err != nil {
		return err
	}

	if hardDelete {

		if err := uow.DB.Unscoped().Delete(&target).Error; err != nil {
			return apperror.NewInternalError("failed to permanently delete user")
		}
	} else {

		target.IsActive = false
		target.DeletedAt = gorm.DeletedAt{Time: time.Now(), Valid: true} // ensure timestamp is set
		if err := uow.DB.Save(&target).Error; err != nil {
			return apperror.NewInternalError("failed to soft delete user")
		}
	}

	return nil
}

func DeleteUserByIDSoftDelete(repo repository.Repository, uow *repository.UnitOfWork, adminID int, userID int) error {
	var admin user.User
	if err := repo.GetByID(uow, uint(adminID), &admin); err != nil {
		return err
	}

	adminData := helper.UserData{
		IsAdmin:  admin.IsAdmin,
		IsActive: admin.IsActive,
	}

	if !helper.IsAuthorizedAdmin(adminData) {
		return apperror.NewAuthError("delete user")
	}

	if adminID == userID {
		return apperror.NewValidationError("user", "admin cannot delete their own account")
	}

	var target user.User
	if err := repo.GetByID(uow, uint(userID), &target); err != nil {
		return err
	}

	if !target.IsActive {
		return apperror.NewValidationError("user", "user is already inactive")
	}

	target.IsActive = false

	if err := uow.DB.Save(&target).Error; err != nil {
		return apperror.NewInternalError("failed to update user as inactive")
	}

	if err := uow.DB.Delete(&target).Error; err != nil {
		return apperror.NewInternalError("failed to soft delete user")
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

func GetAllUsersPaginated(db *gorm.DB, w http.ResponseWriter, r *http.Request, filters ...repository.QueryProcessor) error {
	baseQuery := db.Model(&user.User{})
	var err error

	// Apply filters
	for _, filter := range filters {
		baseQuery, err = filter(baseQuery, &user.User{})
		if err != nil {
			return err
		}
	}

	// Pagination
	pageStr := r.URL.Query().Get("page")
	limitStr := r.URL.Query().Get("limit")
	page, _ := strconv.Atoi(pageStr)
	limit, _ := strconv.Atoi(limitStr)
	if page <= 0 {
		page = 1
	}
	if limit <= 0 {
		limit = 5
	}

	var total int64
	if err := baseQuery.Count(&total).Error; err != nil {
		return err
	}

	var users []user.User
	offset := (page - 1) * limit
	if err := baseQuery.Offset(offset).Limit(limit).Find(&users).Error; err != nil {
		return err
	}

	web.RespondJSON(w, http.StatusOK, map[string]interface{}{
		"data":  users,
		"total": total,
		"page":  page,
		"limit": limit,
	})

	return nil
}
