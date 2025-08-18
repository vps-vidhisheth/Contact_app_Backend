package controller

import (
	"Contact_App/apperror"
	"Contact_App/component/auth"
	"Contact_App/component/contact_detail/service"
	"Contact_App/models/contact_detail"
	"Contact_App/repository"
	"Contact_App/web"
	"context"
	"net/http"
	"strconv"
	"strings"

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

// Get all contact details for a contact with optional filters
func (h *ContactDetailHandler) GetContactDetails(w http.ResponseWriter, r *http.Request) {
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

	// Query params
	detailType := r.URL.Query().Get("type")
	value := r.URL.Query().Get("value")

	// Repository and UnitOfWork
	uow := repository.NewUnitOfWork(h.DB, true) // readonly

	// Base query with filters
	baseQuery := uow.DB.Model(&contact_detail.ContactDetail{}).
		Where("user_id = ?", userID).
		Where("contact_id = ?", contactID)

	if strings.TrimSpace(detailType) != "" {
		baseQuery = baseQuery.Where("type LIKE ?", "%"+detailType+"%")
	}
	if strings.TrimSpace(value) != "" {
		baseQuery = baseQuery.Where("value LIKE ?", "%"+value+"%")
	}

	// Output slice
	var details []*contact_detail.ContactDetail

	// Paginate and respond
	web.Paginate(w, r, uow.DB, &details, baseQuery)
}

func (h *ContactDetailHandler) GetContactDetailByID(w http.ResponseWriter, r *http.Request) {
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

	// Repository and UnitOfWork
	detailRepo := repository.NewGormRepository()
	uow := repository.NewUnitOfWork(h.DB, true) // readonly

	var detail contact_detail.ContactDetail
	filters := []repository.QueryProcessor{
		repository.Filter("user_id = ?", userID),
		repository.Filter("contact_id = ?", contactID),
		repository.Filter("contact_detail_id = ?", detailID),
	}

	err = detailRepo.GetAll(uow, &detail, filters...)
	if err != nil {
		web.RespondError(w, err)
		return
	}

	// If no record found, return 404
	if detail == (contact_detail.ContactDetail{}) {
		web.RespondErrorMessage(w, http.StatusNotFound, "contact detail not found")
		return
	}

	web.RespondJSON(w, http.StatusOK, detail)
}
