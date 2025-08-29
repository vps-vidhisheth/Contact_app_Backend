package modules

import (
	"Contact_App/app"
	"log"
	"net/http"
)

func StartServer(appObj *app.App) {

	SetupRouter(appObj)

	appObj.InitServer(":8080")

	log.Println("Server started at http://localhost:8080")
	if err := appObj.Server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal("Server failed:", err)
	}
}
