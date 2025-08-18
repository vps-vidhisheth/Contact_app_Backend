package modules

import (
	"log"
	"net/http"
)

func StartServer() {
	r := SetupRouter()
	log.Println(" Server started at http://localhost:8080")
	if err := http.ListenAndServe(":8080", r); err != nil {
		log.Fatal("Server failed:", err)
	}
}
