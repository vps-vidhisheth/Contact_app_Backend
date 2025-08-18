package controller

import (
	"Contact_App/component/auth"
	"Contact_App/component/user/service"
	"Contact_App/db"
	"Contact_App/repository"
	"Contact_App/web"
	"encoding/json"
	"net/http"
	"os/user"
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

type loginResponse struct {
	Token string `json:"token"`
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

	user, err := service.Authenticate(db.GetDB(), creds.Email, creds.Password)
	if err != nil {
		web.RespondErrorMessage(w, http.StatusUnauthorized, "Login failed. Please check credentials")
		return
	}

	token, err := service.GenerateJWT(user)
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
	// Extract user ID from URL
	idParam := mux.Vars(r)["userID"]
	id, err := strconv.Atoi(idParam)
	if err != nil {
		web.RespondErrorMessage(w, http.StatusBadRequest, "Invalid user ID")
		return
	}

	uow := repository.NewUnitOfWork(db.GetDB(), true)
	defer uow.Rollback()

	// Optional query params
	fName := r.URL.Query().Get("f_name")
	lName := r.URL.Query().Get("l_name")
	email := r.URL.Query().Get("email")

	// Build filters dynamically
	filters := []repository.QueryProcessor{
		repository.Filter("user_id = ?", id), // always filter by user_id
	}
	if fName != "" {
		filters = append(filters, repository.Filter("f_name LIKE ?", "%"+fName+"%"))
	}
	if lName != "" {
		filters = append(filters, repository.Filter("l_name LIKE ?", "%"+lName+"%"))
	}
	if email != "" {
		filters = append(filters, repository.Filter("email LIKE ?", "%"+email+"%"))
	}

	// Fetch user
	users, err := service.GetAllUsers(repository.NewGormRepository(), uow, filters...)
	if err != nil {
		web.RespondError(w, err)
		return
	}

	// If no user found
	if len(users) == 0 {
		web.RespondErrorMessage(w, http.StatusNotFound, "User not found")
		return
	}

	// Return the first matched user (should be only one because of ID)
	web.RespondJSON(w, http.StatusOK, users[0])
}

func GetAllUsersHandler(w http.ResponseWriter, r *http.Request) {
	uow := repository.NewUnitOfWork(db.GetDB(), true)
	defer uow.Rollback()

	// Base query
	baseQuery := uow.DB.Model(&user.User{})

	// Query params for filtering
	fName := r.URL.Query().Get("f_name")
	lName := r.URL.Query().Get("l_name")
	email := r.URL.Query().Get("email")

	if fName != "" {
		baseQuery = baseQuery.Where("f_name LIKE ?", "%"+fName+"%")
	}
	if lName != "" {
		baseQuery = baseQuery.Where("l_name LIKE ?", "%"+lName+"%")
	}
	if email != "" {
		baseQuery = baseQuery.Where("email LIKE ?", "%"+email+"%")
	}

	// Paginate and respond
	var users []*user.User
	web.Paginate(w, r, uow.DB, &users, baseQuery)
}

func UpdateUserHandler(w http.ResponseWriter, r *http.Request) {
	claims := auth.GetUserClaims(r)
	if claims == nil {
		web.RespondErrorMessage(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	idParam := mux.Vars(r)["userID"]
	id, err := strconv.Atoi(idParam)
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

	adminUser, err := service.GetUserByID(userRepo, uow, claims.UserID)
	if err != nil {
		web.RespondError(w, err)
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
		IsAdmin:  input.IsAdmin,
		Email:    input.Email,
		Password: input.Password,
	}

	updatedUser, err := service.UpdateUserByID(userRepo, uow, adminUser, id, updateData)
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

	err = service.DeleteUserByID(repository.NewGormRepository(), uow, claims.UserID, userID)
	if err != nil {
		web.RespondError(w, err)
		return
	}

	uow.Commit()
	web.RespondJSON(w, http.StatusOK, map[string]string{
		"message": "User deleted permanently",
	})
}
