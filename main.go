package main

// @title Network Scanner API
// @version 1.0
// @description REST API to initiate a network scan and fetch discovered devices
// @host localhost:8080
// @BasePath /
// @schemes http

import (
	"log"
	"net/http"
	"network-scanner/api"
	"network-scanner/config"
	_ "network-scanner/docs"
	"network-scanner/logger"
	"network-scanner/repository"
	"network-scanner/service"

	"github.com/rs/cors"
	httpSwagger "github.com/swaggo/http-swagger"
)

func main() {
	config.LoadConfig()

	appLogger := logger.NewLogger()
	appLogger.Info("Logger initialized")
	dbPath := config.K.String("database.path")
	var repo repository.DeviceRepository
	var err error

	repo, err = repository.NewSQLiteRepository(dbPath, appLogger)
	if err != nil {
		appLogger.Error("Failed to initialize SQLite repo:", err)
		return
	}

	scanner := service.NewScannerService(repo, appLogger)
	scanHandler := api.NewScanHandler(scanner, appLogger)
	deviceHandler := api.NewDeviceHandler(scanner, appLogger)

	mux := http.NewServeMux()
	mux.HandleFunc("/scan", scanHandler.StartScan)
	mux.HandleFunc("/devices", deviceHandler.GetDevices)
	mux.HandleFunc("/clear", deviceHandler.ClearDevices)
	mux.Handle("/swagger/", httpSwagger.WrapHandler)

	corsOptions := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Content-Type"},
		AllowCredentials: true,
	})

	appLogger.Info("Server listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", corsOptions.Handler(mux)))
}
