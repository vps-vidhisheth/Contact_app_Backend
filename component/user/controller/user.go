package controller

import (
	"Contact_App/component/auth"
	"Contact_App/component/user/service"
	"Contact_App/db"
	"Contact_App/repository"
	"Contact_App/web"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

type userInput struct {
	FName    string `json:"f_name"`
	LName    string `json:"l_name"`
	IsAdmin  bool   `json:"is_admin"`
	IsActive bool   `json:"is_active"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func CreateUserHandler(w http.ResponseWriter, r *http.Request) {
	var input userInput
	if err := web.UnmarshalJSON(r, &input); err != nil {
		web.RespondError(w, err)
		return
	}

	uow := repository.NewUnitOfWork(db.GetDB(), false)
	defer uow.Rollback()

	userRepo := repository.NewGormRepository()

	u, err := service.CreateUser(userRepo, uow, input.FName, input.LName, input.IsAdmin, input.IsActive, input.Email, input.Password)
	if err != nil {
		web.RespondError(w, err)
		return
	}

	uow.Commit()
	web.RespondJSON(w, http.StatusCreated, map[string]interface{}{
		"message": "User created successfully",
		"user":    u,
	})
}

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var creds loginRequest
	if err := json.NewDecoder(r.Body).Decode(&creds); err != nil {
		web.RespondErrorMessage(w, http.StatusBadRequest, "Invalid login payload")
		return
	}

	u, err := service.Authenticate(db.GetDB(), creds.Email, creds.Password)
	if err != nil {
		web.RespondErrorMessage(w, http.StatusUnauthorized, "Login failed. Please check credentials")
		return
	}

	token, err := service.GenerateJWT(u)
	if err != nil {
		web.RespondErrorMessage(w, http.StatusInternalServerError, "Failed to generate token")
		return
	}

	web.RespondJSON(w, http.StatusOK, map[string]string{
		"message": "Login successful",
		"token":   token,
	})
}

func GetUserByIDHandler(w http.ResponseWriter, r *http.Request) {
	idParam := mux.Vars(r)["userID"]
	id, err := strconv.Atoi(idParam)
	if err != nil {
		web.RespondErrorMessage(w, http.StatusBadRequest, "Invalid user ID")
		return
	}

	uow := repository.NewUnitOfWork(db.GetDB(), true)
	defer uow.Rollback()

	userRepo := repository.NewGormRepository()

	// Fetch user by ID
	u, err := service.GetUserByID(userRepo, uow, id)
	if err != nil {
		web.RespondErrorMessage(w, http.StatusNotFound, "User not found")
		return
	}

	web.RespondJSON(w, http.StatusOK, u)
}

func GetAllUsersHandler(w http.ResponseWriter, r *http.Request) {
	uow := repository.NewUnitOfWork(db.GetDB(), true)
	defer uow.Rollback()

	userRepo := repository.NewGormRepository()

	// Build filters
	filters := []repository.QueryProcessor{}
	if fName := r.URL.Query().Get("f_name"); fName != "" {
		filters = append(filters, repository.Filter("f_name LIKE ?", "%"+fName+"%"))
	}
	if lName := r.URL.Query().Get("l_name"); lName != "" {
		filters = append(filters, repository.Filter("l_name LIKE ?", "%"+lName+"%"))
	}
	if email := r.URL.Query().Get("email"); email != "" {
		filters = append(filters, repository.Filter("email LIKE ?", "%"+email+"%"))
	}

	users, err := service.GetAllUsers(userRepo, uow, filters...)
	if err != nil {
		web.RespondError(w, err)
		return
	}

	web.RespondJSON(w, http.StatusOK, users)
}
func UpdateUserHandler(w http.ResponseWriter, r *http.Request) {
	claims := auth.GetUserClaims(r)
	if claims == nil || !claims.IsActive {
		web.RespondErrorMessage(w, http.StatusUnauthorized, "Unauthorized or inactive user")
		return
	}

	idParam := mux.Vars(r)["userID"]
	userID, err := strconv.Atoi(idParam)
	if err != nil {
		web.RespondErrorMessage(w, http.StatusBadRequest, "Invalid user ID")
		return
	}

	var input userInput
	if err := web.UnmarshalJSON(r, &input); err != nil {
		web.RespondError(w, err)
		return
	}

	uow := repository.NewUnitOfWork(db.GetDB(), false)
	defer uow.Rollback()

	userRepo := repository.NewGormRepository()

	// Admins can update any user; non-admins can update only themselves
	if !claims.IsAdmin && claims.UserID != userID {
		web.RespondErrorMessage(w, http.StatusForbidden, "You can only update your own profile")
		return
	}

	updateData := &struct {
		FName    string `json:"f_name"`
		LName    string `json:"l_name"`
		IsAdmin  bool   `json:"is_admin"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}{
		FName:    input.FName,
		LName:    input.LName,
		Email:    input.Email,
		Password: input.Password,
	}

	// Only admins can update the IsAdmin field
	if claims.IsAdmin {
		updateData.IsAdmin = input.IsAdmin
	}

	updatedUser, err := service.UpdateUserByID(userRepo, uow, claims, userID, updateData)
	if err != nil {
		web.RespondError(w, err)
		return
	}

	uow.Commit()
	web.RespondJSON(w, http.StatusOK, map[string]interface{}{
		"message": "User updated successfully",
		"user":    updatedUser,
	})
}

func DeleteUserHandler(w http.ResponseWriter, r *http.Request) {
	claims := auth.GetUserClaims(r)
	if claims == nil {
		web.RespondErrorMessage(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	userIDStr := mux.Vars(r)["userID"]
	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		web.RespondErrorMessage(w, http.StatusBadRequest, "Invalid user ID")
		return
	}

	uow := repository.NewUnitOfWork(db.GetDB(), false)
	defer uow.Rollback()

	userRepo := repository.NewGormRepository()

	err = service.DeleteUserByIDSoftDelete(userRepo, uow, claims.UserID, userID)
	if err != nil {
		web.RespondError(w, err)
		return
	}

	uow.Commit()
	web.RespondJSON(w, http.StatusOK, map[string]string{
		"message": "User deleted successfully",
	})
}
