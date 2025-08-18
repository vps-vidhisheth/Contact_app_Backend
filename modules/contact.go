package modules

import (
	"Contact_App/component/auth"
	contactCtrl "Contact_App/component/contact/controller"

	"github.com/gorilla/mux"
)

func RegisterContactRoutes(r *mux.Router) {
	c := r.PathPrefix("/users/{userID}/contacts").Subrouter()

	c.HandleFunc("", contactCtrl.CreateContactHandler).Methods("POST")
	c.HandleFunc("", contactCtrl.GetContactsHandler).Methods("GET")
	c.HandleFunc("/{contactID}", contactCtrl.GetContactByIDHandler).Methods("GET")
	c.HandleFunc("/{contactID}", contactCtrl.UpdateContactHandler).Methods("PUT")
	c.HandleFunc("/{contactID}", contactCtrl.DeleteContactHandler).Methods("DELETE")
	c.HandleFunc("/with-details", contactCtrl.AddContactWithDetailsHandler).Methods("POST")

	c.Use(auth.MiddlewareStaffActive)
}
