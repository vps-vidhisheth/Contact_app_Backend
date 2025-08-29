package app

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"Contact_App/component/auth"
	contactController "Contact_App/component/contact/controller"
	"Contact_App/component/contact/service"
	contactDetailController "Contact_App/component/contact_detail/controller"
	userController "Contact_App/component/user/controller"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"gorm.io/gorm"
)

type App struct {
	sync.Mutex
	Router *mux.Router
	DB     *gorm.DB
	Server *http.Server
	WG     *sync.WaitGroup
}

func NewApp(db *gorm.DB) *App {
	wg := &sync.WaitGroup{}
	app := &App{
		Router: mux.NewRouter().StrictSlash(true),
		DB:     db,
		WG:     wg,
	}

	app.registerRoutes()
	return app
}

func (app *App) InitServer(port string) {
	headers := handlers.AllowedHeaders([]string{"Content-Type", "X-Total-Count", "token", "Authorization"})
	methods := handlers.AllowedMethods([]string{"GET", "POST", "PUT", "DELETE", "OPTIONS"})
	origins := handlers.AllowedOrigins([]string{"*"})

	app.Server = &http.Server{
		Addr:    port,
		Handler: handlers.CORS(headers, methods, origins)(app.Router),
	}
}

func (app *App) StartServer() {
	fmt.Printf("Server running on %s...\n", app.Server.Addr)
	if err := app.Server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("ListenAndServe error: %v", err)
	}
}

func (app *App) Stop() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	sqlDB, err := app.DB.DB()
	if err == nil {
		sqlDB.Close()
		fmt.Println("DB Closed")
	}

	if err := app.Server.Shutdown(ctx); err != nil {
		fmt.Printf("Failed to stop server gracefully: %v\n", err)
		return
	}
	fmt.Println("Server Shutdown Gracefully.")
}

func (app *App) registerRoutes() {
	api := app.Router.PathPrefix("/api/v1").Subrouter()

	api.Use(auth.AttachUserToContextIfTokenValid)

	uHandler := userController.NewUserHandler(app.DB)
	cService := service.NewContactService()
	cController := contactController.NewContactController(cService)
	cdHandler := contactDetailController.NewContactDetailHandler(app.DB)

	uHandler.RegisterRoutes(api)
	cController.RegisterRoutes(api)
	cdHandler.RegisterRoutes(api)
}
