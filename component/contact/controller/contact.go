package controller

import (
	"Contact_App/apperror"
	"Contact_App/component/auth"
	"Contact_App/component/contact/service"
	detailservice "Contact_App/component/contact_detail/service"
	"Contact_App/db"
	"Contact_App/web"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
)

// CREATE CONTACT
// CREATE CONTACT WITH DETAILS
func CreateContactHandler(w http.ResponseWriter, r *http.Request) {
	claims := auth.GetUserClaims(r)
	if claims == nil {
		web.RespondErrorMessage(w, http.StatusUnauthorized, "unauthorized user")
		return
	}

	userIDStr := mux.Vars(r)["userID"]
	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		web.RespondErrorMessage(w, http.StatusBadRequest, "invalid user ID in path")
		return
	}

	if claims.UserID != userID {
		web.RespondErrorMessage(w, http.StatusForbidden, "you can only create contacts for yourself")
		return
	}

	// Accept both contact info and optional details
	var input struct {
		FName    string `json:"first_name"`
		LName    string `json:"last_name"`
		IsActive bool   `json:"is_active"`
		Details  []struct {
			Type  string `json:"type"`
			Value string `json:"value"`
		} `json:"details"`
	}

	if err := web.UnmarshalJSON(r, &input); err != nil {
		web.RespondError(w, err)
		return
	}

	// Create contact
	contact, err := service.CreateContact(userID, input.FName, input.LName)
	if err != nil {
		web.RespondError(w, err)
		return
	}

	// Add details if provided
	for _, d := range input.Details {
		_, err := detailservice.AddDetailToContact(db.GetDB(), userID, contact.ContactID, d.Type, d.Value)
		if err != nil {
			web.RespondError(w, err)
			return
		}
	}

	web.RespondJSON(w, http.StatusCreated, contact)
}

// GET ALL CONTACTS (with optional filters: f_name, l_name, phone)
func GetContactsHandler(w http.ResponseWriter, r *http.Request) {
	userIDStr := mux.Vars(r)["userID"]
	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		web.RespondErrorMessage(w, http.StatusBadRequest, "invalid user ID")
		return
	}

	// Read optional query params
	filters := map[string]string{
		"f_name": r.URL.Query().Get("f_name"),
		"l_name": r.URL.Query().Get("l_name"),
		"phone":  r.URL.Query().Get("phone"),
	}

	contacts, err := service.GetContactsWithDetails(userID, filters)
	if err != nil {
		web.RespondError(w, err)
		return
	}

	web.RespondJSON(w, http.StatusOK, contacts)
}

// GET CONTACT BY ID
func GetContactByIDHandler(w http.ResponseWriter, r *http.Request) {
	userIDStr := mux.Vars(r)["userID"]
	contactIDStr := mux.Vars(r)["contactID"]

	userID, err1 := strconv.Atoi(userIDStr)
	contactID, err2 := strconv.Atoi(contactIDStr)

	if err1 != nil || err2 != nil {
		web.RespondErrorMessage(w, http.StatusBadRequest, "invalid path parameters")
		return
	}

	contact, err := service.GetContactByIDWithDetails(userID, contactID)
	if err != nil {
		if _, ok := err.(*apperror.NotFoundError); ok {
			web.RespondErrorMessage(w, http.StatusNotFound, "contact not found")
		} else {
			web.RespondError(w, err)
		}
		return
	}

	web.RespondJSON(w, http.StatusOK, contact)
}

// UPDATE CONTACT BY ID
func UpdateContactHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userIDStr, ok1 := vars["userID"]
	contactIDStr, ok2 := vars["contactID"]

	if !ok1 || !ok2 {
		apperror.HandleBadRequest(w, "Missing path parameters: userID or contactID")
		return
	}

	userID, err1 := strconv.Atoi(userIDStr)
	contactID, err2 := strconv.Atoi(contactIDStr)
	if err1 != nil || err2 != nil {
		apperror.HandleBadRequest(w, "Invalid path parameters: must be integers")
		return
	}

	// Define input structure
	type ContactDetailInput struct {
		Type  string `json:"type"`
		Value string `json:"value"`
	}

	type UpdateContactInput struct {
		FName    string               `json:"first_name"`
		LName    string               `json:"last_name"`
		IsActive bool                 `json:"is_active"`
		Details  []ContactDetailInput `json:"details"`
	}

	// Decode JSON payload
	var input UpdateContactInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		apperror.HandleBadRequest(w, "Invalid JSON payload")
		return
	}

	input.FName = strings.TrimSpace(input.FName)
	input.LName = strings.TrimSpace(input.LName)

	if input.FName == "" || input.LName == "" {
		apperror.HandleBadRequest(w, "First name and Last name are required")
		return
	}

	// Update contact fields
	if err := service.UpdateContactByID(userID, contactID, "fname", input.FName); err != nil {
		apperror.HandleError(w, err)
		return
	}
	if err := service.UpdateContactByID(userID, contactID, "lname", input.LName); err != nil {
		apperror.HandleError(w, err)
		return
	}
	if err := service.UpdateContactByID(userID, contactID, "is_active", input.IsActive); err != nil {
		apperror.HandleError(w, err)
		return
	}

	// Update contact details
	for _, d := range input.Details {
		d.Type = strings.TrimSpace(d.Type)
		d.Value = strings.TrimSpace(d.Value)
		if d.Type == "" || d.Value == "" {
			continue // skip empty details
		}

		if err := service.AddOrUpdateContactDetail(contactID, d.Type, d.Value); err != nil {
			apperror.HandleError(w, err)
			return
		}
	}

	// Success response
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Contact updated successfully",
	})
}

// DELETE CONTACT BY ID
func DeleteContactHandler(w http.ResponseWriter, r *http.Request) {
	userIDStr := mux.Vars(r)["userID"]
	contactIDStr := mux.Vars(r)["contactID"]

	userID, err1 := strconv.Atoi(userIDStr)
	contactID, err2 := strconv.Atoi(contactIDStr)

	if err1 != nil || err2 != nil {
		web.RespondErrorMessage(w, http.StatusBadRequest, "invalid path parameters")
		return
	}

	if err := service.DeleteContactByID(userID, contactID); err != nil {
		web.RespondError(w, err)
		return
	}

	web.RespondJSON(w, http.StatusOK, map[string]string{"message": "contact deleted"})
}

// ADD CONTACT WITH DETAILS
func AddContactWithDetailsHandler(w http.ResponseWriter, r *http.Request) {
	claims := auth.GetUserClaims(r)
	if claims == nil {
		web.RespondErrorMessage(w, http.StatusUnauthorized, "unauthorized user")
		return
	}

	userIDStr := mux.Vars(r)["userID"]
	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		web.RespondErrorMessage(w, http.StatusBadRequest, "invalid user ID in path")
		return
	}

	if claims.UserID != userID {
		web.RespondErrorMessage(w, http.StatusForbidden, "you can only add contacts for yourself")
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

	if err := web.UnmarshalJSON(r, &input); err != nil {
		web.RespondError(w, err)
		return
	}

	contact, err := service.CreateContact(userID, input.FName, input.LName)
	if err != nil {
		web.RespondError(w, err)
		return
	}

	for _, d := range input.Details {
		_, err := detailservice.AddDetailToContact(db.GetDB(), userID, contact.ContactID, d.Type, d.Value)
		if err != nil {
			web.RespondError(w, err)
			return
		}
	}

	web.RespondJSON(w, http.StatusCreated, contact)
}
