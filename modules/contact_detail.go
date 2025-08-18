package modules

import (
	"Contact_App/component/auth"
	cdCtrl "Contact_App/component/contact_detail/controller"

	"github.com/gorilla/mux"
	"gorm.io/gorm"
)

func RegisterContactDetailRoutes(r *mux.Router, db *gorm.DB) {
	handler := &cdCtrl.ContactDetailHandler{DB: db}

	cd := r.PathPrefix("/user/{userID}/contacts/{contact_id}/details").Subrouter()

	cd.HandleFunc("", handler.AddContactDetail).Methods("POST")
	cd.HandleFunc("/{detail_id}", handler.UpdateContactDetail).Methods("PUT")
	cd.HandleFunc("/{detail_id}", handler.DeleteContactDetail).Methods("DELETE")

	cd.Use(auth.MiddlewareStaffActive)
}
