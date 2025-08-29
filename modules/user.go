package modules

import (
	"Contact_App/app"
	userCtrl "Contact_App/component/user/controller"
)

func RegisterUserRoutes(appObj *app.App) {

	userHandler := userCtrl.NewUserHandler(appObj.DB)

	userHandler.RegisterRoutes(appObj.Router)
}
