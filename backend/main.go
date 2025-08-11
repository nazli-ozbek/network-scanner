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

	httpSwagger "github.com/swaggo/http-swagger"
)

func main() {
	repo := repository.NewInMemoryRepository()
	scanner := service.NewScannerService(repo)
	handler := api.NewHandler(scanner)

	http.HandleFunc("/scan", handler.StartScan)
	http.HandleFunc("/devices", handler.GetDevices)
	http.HandleFunc("/clear", handler.ClearDevices)

	http.Handle("/swagger/", httpSwagger.WrapHandler)

	log.Println("Server started at http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
