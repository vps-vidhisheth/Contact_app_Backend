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
	contactID, err := strconv.Atoi(vars["contact_id"])
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
	contactID, _ := strconv.Atoi(vars["contact_id"])
	detailID, _ := strconv.Atoi(vars["detail_id"])

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
	userID, _ := extractUserID(r.Context())
	vars := mux.Vars(r)
	contactID, _ := strconv.Atoi(vars["contact_id"])
	detailID, _ := strconv.Atoi(vars["detail_id"])

	if err := service.DeleteDetailByID(h.DB, userID, contactID, detailID); err != nil {
		web.RespondError(w, err)
		return
	}

	web.RespondJSON(w, http.StatusOK, map[string]string{"message": "Contact detail deleted successfully"})
}

func (h *ContactDetailHandler) GetContactDetails(w http.ResponseWriter, r *http.Request) {
	userID, _ := extractUserID(r.Context())
	contactID, _ := strconv.Atoi(mux.Vars(r)["contact_id"])

	detailType := r.URL.Query().Get("type")
	value := r.URL.Query().Get("value")

	uow := repository.NewUnitOfWork(h.DB, true)
	defer uow.Rollback()

	baseQuery := uow.DB.Model(&contact_detail.ContactDetail{}).
		Where("user_id = ?", uint(userID)).
		Where("contact_id = ?", uint(contactID))

	if strings.TrimSpace(detailType) != "" {
		baseQuery = baseQuery.Where("type LIKE ?", "%"+detailType+"%")
	}
	if strings.TrimSpace(value) != "" {
		baseQuery = baseQuery.Where("value LIKE ?", "%"+value+"%")
	}

	var details []*contact_detail.ContactDetail
	web.Paginate(w, r, uow.DB, &details, baseQuery)
}

func (h *ContactDetailHandler) GetContactDetailByID(w http.ResponseWriter, r *http.Request) {
	userID, _ := extractUserID(r.Context())
	vars := mux.Vars(r)
	contactID, _ := strconv.Atoi(vars["contact_id"])
	detailID, _ := strconv.Atoi(vars["detail_id"])

	detailRepo := repository.NewGormRepository()
	uow := repository.NewUnitOfWork(h.DB, true)
	defer uow.Rollback()

	var detail contact_detail.ContactDetail
	filters := []repository.QueryProcessor{
		repository.Filter("user_id = ?", uint(userID)),
		repository.Filter("contact_id = ?", uint(contactID)),
		repository.Filter("contact_details_id = ?", uint(detailID)),
	}

	err := detailRepo.GetAll(uow, &detail, filters...)
	if err != nil {
		web.RespondError(w, err)
		return
	}

	if detail == (contact_detail.ContactDetail{}) {
		web.RespondErrorMessage(w, http.StatusNotFound, "contact detail not found")
		return
	}

	web.RespondJSON(w, http.StatusOK, detail)
}

func (h *ContactDetailHandler) RegisterRoutes(router *mux.Router) {
	router.HandleFunc("/users/{user_id}/contacts/{contact_id}/details", h.GetContactDetails).Methods("GET")
	router.HandleFunc("/users/{user_id}/contacts/{contact_id}/details", h.AddContactDetail).Methods("POST")
	router.HandleFunc("/users/{user_id}/contacts/{contact_id}/details/{detail_id}", h.GetContactDetailByID).Methods("GET")
	router.HandleFunc("/users/{user_id}/contacts/{contact_id}/details/{detail_id}", h.UpdateContactDetail).Methods("PUT")
	router.HandleFunc("/users/{user_id}/contacts/{contact_id}/details/{detail_id}", h.DeleteContactDetail).Methods("DELETE")
}
