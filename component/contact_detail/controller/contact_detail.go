package controller

import (
	"Contact_App/apperror"
	"Contact_App/component/auth"
	"Contact_App/component/contact_detail/service"
	"Contact_App/web"
	"context"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"gorm.io/gorm"
)

type ContactDetailHandler struct {
	DB *gorm.DB
}

func NewContactDetailHandler(db *gorm.DB) *ContactDetailHandler {
	return &ContactDetailHandler{DB: db}
}

func extractUserID(ctx context.Context) (int, error) {
	claims, ok := ctx.Value(auth.GetUserContextKey()).(*auth.Claims)
	if !ok || claims == nil {
		return 0, apperror.NewUnauthorized("userID")
	}
	return claims.UserID, nil
}

func (h *ContactDetailHandler) AddContactDetail(w http.ResponseWriter, r *http.Request) {
	userID, err := extractUserID(r.Context())
	if err != nil {
		web.RespondError(w, err)
		return
	}

	vars := mux.Vars(r)
	contactIDStr, ok := vars["contact_id"]
	if !ok {
		web.RespondError(w, apperror.NewValidationError("contact_id", "missing in URL path"))
		return
	}

	contactID, err := strconv.Atoi(contactIDStr)
	if err != nil {
		web.RespondError(w, apperror.NewValidationError("contact_id", "must be an integer"))
		return
	}

	var req web.CreateDetailRequest
	if err := web.UnmarshalJSON(r, &req); err != nil {
		web.RespondError(w, err)
		return
	}

	detail, err := service.AddDetailToContact(h.DB, userID, contactID, req.Type, req.Value)
	if err != nil {
		web.RespondError(w, err)
		return
	}

	web.RespondJSON(w, http.StatusCreated, detail)
}

func (h *ContactDetailHandler) UpdateContactDetail(w http.ResponseWriter, r *http.Request) {
	userID, err := extractUserID(r.Context())
	if err != nil {
		web.RespondError(w, err)
		return
	}

	vars := mux.Vars(r)
	contactIDStr, contactOK := vars["contact_id"]
	detailIDStr, detailOK := vars["detail_id"]

	if !contactOK || !detailOK {
		web.RespondError(w, apperror.NewValidationError("path", "contact_id or detail_id missing"))
		return
	}

	contactID, err := strconv.Atoi(contactIDStr)
	if err != nil {
		web.RespondError(w, apperror.NewValidationError("contact_id", "must be an integer"))
		return
	}

	detailID, err := strconv.Atoi(detailIDStr)
	if err != nil {
		web.RespondError(w, apperror.NewValidationError("detail_id", "must be an integer"))
		return
	}

	var req web.UpdateDetailRequest
	if err := web.UnmarshalJSON(r, &req); err != nil {
		web.RespondError(w, err)
		return
	}

	input := service.UpdateDetailInput{
		Email: req.Email,
		Phone: req.Phone,
	}

	if err := service.UpdateDetailByID(h.DB, userID, contactID, detailID, input); err != nil {
		web.RespondError(w, err)
		return
	}

	web.RespondJSON(w, http.StatusOK, map[string]string{"message": "Contact detail updated successfully"})
}

func (h *ContactDetailHandler) DeleteContactDetail(w http.ResponseWriter, r *http.Request) {
	userID, err := extractUserID(r.Context())
	if err != nil {
		web.RespondError(w, err)
		return
	}

	vars := mux.Vars(r)
	contactIDStr, contactOK := vars["contact_id"]
	detailIDStr, detailOK := vars["detail_id"]

	if !contactOK || !detailOK {
		web.RespondError(w, apperror.NewValidationError("path", "contact_id or detail_id missing"))
		return
	}

	contactID, err := strconv.Atoi(contactIDStr)
	if err != nil {
		web.RespondError(w, apperror.NewValidationError("contact_id", "must be an integer"))
		return
	}

	detailID, err := strconv.Atoi(detailIDStr)
	if err != nil {
		web.RespondError(w, apperror.NewValidationError("detail_id", "must be an integer"))
		return
	}

	if err := service.DeleteDetailByID(h.DB, userID, contactID, detailID); err != nil {
		web.RespondError(w, err)
		return
	}

	web.RespondJSON(w, http.StatusOK, map[string]string{"message": "Contact detail deleted successfully"})
}
