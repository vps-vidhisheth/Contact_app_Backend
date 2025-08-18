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

	var input struct {
		FName string `json:"first_name"`
		LName string `json:"last_name"`
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

	web.RespondJSON(w, http.StatusCreated, contact)
}

func GetContactsHandler(w http.ResponseWriter, r *http.Request) {
	userIDStr := mux.Vars(r)["userID"]
	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		web.RespondErrorMessage(w, http.StatusBadRequest, "invalid user ID")
		return
	}

	contacts, err := service.GetContacts(userID)
	if err != nil {
		web.RespondError(w, err)
		return
	}

	web.RespondJSON(w, http.StatusOK, contacts)
}

func GetContactByIDHandler(w http.ResponseWriter, r *http.Request) {
	userID, err1 := strconv.Atoi(mux.Vars(r)["userID"])
	contactID, err2 := strconv.Atoi(mux.Vars(r)["contact_id"])
	if err1 != nil || err2 != nil {
		web.RespondErrorMessage(w, http.StatusBadRequest, "invalid path parameters")
		return
	}

	contact, err := service.GetContactByID(userID, contactID)
	if err != nil {
		web.RespondError(w, err)
		return
	}

	web.RespondJSON(w, http.StatusOK, contact)
}

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

	var input service.UpdateFieldInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		apperror.HandleBadRequest(w, "Invalid JSON payload")
		return
	}

	input.Field = strings.TrimSpace(input.Field)
	input.Value = strings.TrimSpace(input.Value)
	if input.Field == "" || input.Value == "" {
		apperror.HandleBadRequest(w, "Field and value are required and cannot be empty")
		return
	}

	if err := service.UpdateContactByID(userID, contactID, input.Field, input.Value); err != nil {
		apperror.HandleError(w, err)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Contact updated successfully",
	})
}

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
