package modules

import (
	"Contact_App/app"
	contactCtrl "Contact_App/component/contact/controller"
	"Contact_App/component/contact/service"
)

func RegisterContactRoutes(appObj *app.App) {

	contactService := service.NewContactService()

	contactController := contactCtrl.NewContactController(contactService)

	contactController.RegisterRoutes(appObj.Router)
}
