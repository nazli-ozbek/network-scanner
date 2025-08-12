// @title Network Scanner API
// @version 1.0
// @description REST API to initiate a network scan and fetch discovered devices
// @host localhost:8080
// @BasePath /
// @schemes http

package main

import (
	"log"
	"net/http"
	"network-scanner/api"
	_ "network-scanner/docs"
	"network-scanner/repository"
	"network-scanner/service"

	"github.com/rs/cors"
	httpSwagger "github.com/swaggo/http-swagger"
)

func main() {
	repo := repository.NewInMemoryRepository()
	scanner := service.NewScannerService(repo)
	handler := api.NewHandler(scanner)

	mux := http.NewServeMux()
	mux.HandleFunc("/scan", handler.StartScan)
	mux.HandleFunc("/devices", handler.GetDevices)
	mux.HandleFunc("/clear", handler.ClearDevices)

	mux.Handle("/swagger/", httpSwagger.WrapHandler)
	corsOptions := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Content-Type"},
		AllowCredentials: true,
	})

	corsHandler := corsOptions.Handler(mux)

	log.Println("Server listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", corsHandler))
}
