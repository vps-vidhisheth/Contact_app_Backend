package modules

import (
	"Contact_App/app"
	cdCtrl "Contact_App/component/contact_detail/controller"
)

func RegisterContactDetailRoutes(appObj *app.App) {

	contactDetailHandler := cdCtrl.NewContactDetailHandler(appObj.DB)

	contactDetailHandler.RegisterRoutes(appObj.Router)
}
