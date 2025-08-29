package modules

import (
	"Contact_App/app"
	"Contact_App/component/auth"
	"log"

	"github.com/gorilla/mux"
)

func SetupRouter(appObj *app.App) {
	appObj.Router.Use(auth.AttachUserToContextIfTokenValid)

	RegisterUserRoutes(appObj)
	RegisterContactRoutes(appObj)
	RegisterContactDetailRoutes(appObj)

	if err := appObj.Router.Walk(func(route *mux.Route, router *mux.Router, ancestors []*mux.Route) error {
		path, _ := route.GetPathTemplate()
		methods, _ := route.GetMethods()
		log.Printf("Registered Route: %-6v %s\n", methods, path)
		return nil
	}); err != nil {
		log.Printf("Error walking routes: %v\n", err)
	}
}
