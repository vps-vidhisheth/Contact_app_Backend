package controller

import (
	"Contact_App/apperror"
	"Contact_App/component/auth"
	"Contact_App/component/contact/service"
	"Contact_App/db"
	"Contact_App/repository"
	"Contact_App/web"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
)

type ContactController struct {
	Service *service.ContactService
}

func NewContactController(svc *service.ContactService) *ContactController {
	return &ContactController{Service: svc}
}

func (c *ContactController) CreateContactHandler(w http.ResponseWriter, r *http.Request) {
	claims := auth.GetUserClaims(r)
	if claims == nil {
		apperror.HandleUnauthorized(w, "missing or invalid token")
		return
	}

	userID64, err := strconv.ParseUint(mux.Vars(r)["userID"], 10, 64)
	if err != nil {
		apperror.HandleBadRequest(w, "invalid userID")
		return
	}
	userID := uint(userID64)

	if claims.UserID != int(userID) {
		http.Error(w, "Forbidden: cannot create contacts for another user", http.StatusForbidden)
		return
	}

	var input struct {
		FName   string `json:"first_name"`
		LName   string `json:"last_name"`
		Details []struct {
			Type  string `json:"type"`
			Value string `json:"value"`
		} `json:"details"`
	}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		apperror.HandleBadRequest(w, "invalid JSON")
		return
	}

	input.FName = strings.TrimSpace(input.FName)
	input.LName = strings.TrimSpace(input.LName)
	if input.FName == "" || input.LName == "" {
		apperror.HandleBadRequest(w, "first_name and last_name required")
		return
	}

	uow := repository.NewUnitOfWork(db.GetDB(), false)
	defer uow.Rollback()

	contactObj, err := c.Service.CreateContactWithUOW(uow, userID, input.FName, input.LName)
	if err != nil {
		apperror.HandleError(w, err)
		return
	}

	for _, d := range input.Details {
		if d.Type == "" || d.Value == "" {
			continue
		}
		if err := c.Service.AddOrUpdateContactDetail(uow, userID, contactObj.ContactID, d.Type, d.Value); err != nil {
			apperror.HandleError(w, err)
			return
		}
	}

	uow.Commit()

	web.RespondJSON(w, http.StatusCreated, contactObj)
}

func (c *ContactController) GetContactsHandler(w http.ResponseWriter, r *http.Request) {
	claims := auth.GetUserClaims(r)
	if claims == nil {
		apperror.HandleUnauthorized(w, "missing or invalid token")
		return
	}

	userID64, err := strconv.ParseUint(mux.Vars(r)["userID"], 10, 64)
	if err != nil {
		apperror.HandleBadRequest(w, "invalid userID")
		return
	}
	userID := uint(userID64)

	if claims.UserID != int(userID) {
		http.Error(w, "Forbidden: cannot fetch contacts of another user", http.StatusForbidden)
		return
	}

	filters := map[string]string{
		"f_name": r.URL.Query().Get("f_name"),
		"l_name": r.URL.Query().Get("l_name"),
		"phone":  r.URL.Query().Get("phone"),
	}

	contacts, err := c.Service.GetContactsWithDetails(userID, filters)
	if err != nil {
		apperror.HandleError(w, err)
		return
	}

	web.RespondJSON(w, http.StatusOK, contacts)
}

// GET /users/{userID}/contacts/{contactID}
func (c *ContactController) GetContactByIDHandler(w http.ResponseWriter, r *http.Request) {
	claims := auth.GetUserClaims(r)
	if claims == nil {
		apperror.HandleUnauthorized(w, "missing or invalid token")
		return
	}

	userID64, _ := strconv.ParseUint(mux.Vars(r)["userID"], 10, 64)
	contactID64, _ := strconv.ParseUint(mux.Vars(r)["contactID"], 10, 64)
	userID := uint(userID64)
	contactID := uint(contactID64)

	if claims.UserID != int(userID) {
		http.Error(w, "Forbidden: cannot fetch another user's contact", http.StatusForbidden)
		return
	}

	contactObj, err := c.Service.GetContactByIDWithDetails(userID, contactID)
	if err != nil {
		apperror.HandleError(w, err)
		return
	}

	web.RespondJSON(w, http.StatusOK, contactObj)
}

func (c *ContactController) UpdateContactHandler(w http.ResponseWriter, r *http.Request) {
	claims := auth.GetUserClaims(r)
	if claims == nil {
		apperror.HandleUnauthorized(w, "missing or invalid token")
		return
	}

	userID64, _ := strconv.ParseUint(mux.Vars(r)["userID"], 10, 64)
	contactID64, _ := strconv.ParseUint(mux.Vars(r)["contactID"], 10, 64)
	userID := uint(userID64)
	contactID := uint(contactID64)

	if claims.UserID != int(userID) {
		http.Error(w, "Forbidden: cannot update another user's contact", http.StatusForbidden)
		return
	}

	var updates map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
		apperror.HandleBadRequest(w, "invalid JSON")
		return
	}

	if err := c.Service.UpdateContactByID(userID, contactID, updates); err != nil {
		apperror.HandleError(w, err)
		return
	}

	web.RespondJSON(w, http.StatusOK, map[string]string{"message": "contact updated"})
}

// DELETE /users/{userID}/contacts/{contactID}
func (c *ContactController) DeleteContactHandler(w http.ResponseWriter, r *http.Request) {
	claims := auth.GetUserClaims(r)
	if claims == nil {
		apperror.HandleUnauthorized(w, "missing or invalid token")
		return
	}

	userID64, _ := strconv.ParseUint(mux.Vars(r)["userID"], 10, 64)
	contactID64, _ := strconv.ParseUint(mux.Vars(r)["contactID"], 10, 64)
	userID := uint(userID64)
	contactID := uint(contactID64)

	if claims.UserID != int(userID) {
		http.Error(w, "Forbidden: cannot delete another user's contact", http.StatusForbidden)
		return
	}

	if err := c.Service.DeleteContactByID(userID, contactID); err != nil {
		apperror.HandleError(w, err)
		return
	}

	web.RespondJSON(w, http.StatusOK, map[string]string{"message": "contact deleted"})
}

func (c *ContactController) RegisterRoutes(router *mux.Router) {
	router.HandleFunc("/users/{userID}/contacts", c.CreateContactHandler).Methods("POST")
	router.HandleFunc("/users/{userID}/contacts", c.GetContactsHandler).Methods("GET")
	router.HandleFunc("/users/{userID}/contacts/{contactID}", c.GetContactByIDHandler).Methods("GET")
	router.HandleFunc("/users/{userID}/contacts/{contactID}", c.UpdateContactHandler).Methods("PUT")
	router.HandleFunc("/users/{userID}/contacts/{contactID}", c.DeleteContactHandler).Methods("DELETE")
}
