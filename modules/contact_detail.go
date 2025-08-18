package modules

import (
	cdCtrl "Contact_App/component/contact_detail/controller"

	"github.com/gorilla/mux"
	"gorm.io/gorm"
)

func RegisterContactDetailRoutes(r *mux.Router, db *gorm.DB) {
	handler := &cdCtrl.ContactDetailHandler{DB: db}

	r.HandleFunc("/users/{user_id}/contacts/{contact_id}/details", handler.GetContactDetails).Methods("GET")
	r.HandleFunc("/users/{user_id}/contacts/{contact_id}/details", handler.AddContactDetail).Methods("POST")

	r.HandleFunc("/users/{user_id}/contacts/{contact_id}/details/{detail_id}", handler.GetContactDetailByID).Methods("GET")
	r.HandleFunc("/users/{user_id}/contacts/{contact_id}/details/{detail_id}", handler.UpdateContactDetail).Methods("PUT")
	r.HandleFunc("/users/{user_id}/contacts/{contact_id}/details/{detail_id}", handler.DeleteContactDetail).Methods("DELETE")
}
