package modules

import (
	"Contact_App/component/auth"
	"Contact_App/db"
	"log"

	"github.com/gorilla/mux"
)

func SetupRouter() *mux.Router {
	db.InitDB()

	r := mux.NewRouter().StrictSlash(true)

	r.Use(auth.AttachUserToContextIfTokenValid)

	RegisterUserRoutes(r)
	RegisterContactRoutes(r)
	RegisterContactDetailRoutes(r, db.GetDB())

	err := r.Walk(func(route *mux.Route, router *mux.Router, ancestors []*mux.Route) error {
		path, _ := route.GetPathTemplate()
		methods, _ := route.GetMethods()
		log.Printf(" Registered Route: %-6s %s\n", methods, path)
		return nil
	})
	if err != nil {
		log.Printf(" Error walking routes: %v\n", err)
	}

	return r
}
