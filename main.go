package main

// @title Network Scanner API
// @version 1.0
// @description REST API to initiate a network scan and fetch discovered devices
// @host localhost:8080
// @BasePath /
// @schemes http

import (
	"database/sql"
	"log"
	"net/http"
	"time"

	"network-scanner/api"
	"network-scanner/config"
	_ "network-scanner/docs"
	"network-scanner/logger"
	"network-scanner/repository"
	"network-scanner/service"

	"github.com/gorilla/mux"
	_ "github.com/mattn/go-sqlite3"
	"github.com/rs/cors"
	httpSwagger "github.com/swaggo/http-swagger"
)

func main() {
	config.LoadConfig()

	appLogger := logger.NewLogger()
	appLogger.Info("Logger initialized")

	dbPath := config.K.String("database.path")

	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		appLogger.Error("failed to open sqlite db:", err)
		return
	}

	deviceRepo, err := repository.NewSQLiteRepositoryWithDB(db, appLogger)
	if err != nil {
		appLogger.Error("Failed to initialize device repo:", err)
		return
	}

	scanner := service.NewScannerService(deviceRepo, appLogger)
	scanner.StartStatusPolling(5 * time.Second)
	scanHandler := api.NewScanHandler(scanner, appLogger)
	deviceHandler := api.NewDeviceHandler(scanner, appLogger)

	rangeRepo := repository.NewSQLiteIPRangeRepository(db, appLogger)
	rangeService := service.NewRangeService(rangeRepo)
	rangeHandler := api.NewRangeHandler(rangeService, appLogger)

	r := mux.NewRouter()

	r.HandleFunc("/scan", scanHandler.StartScan).Methods("POST")
	r.HandleFunc("/devices", deviceHandler.GetDevices).Methods("GET")
	r.HandleFunc("/clear", deviceHandler.ClearDevices).Methods("DELETE")

	r.HandleFunc("/devices/search", deviceHandler.SearchDevices).Methods("GET")
	r.HandleFunc("/devices/{id}", deviceHandler.GetDeviceByID).Methods("GET")
	r.HandleFunc("/devices/{id}/tags", deviceHandler.AddTag).Methods("POST")
	r.HandleFunc("/devices/{id}/tags", deviceHandler.RemoveTag).Methods("DELETE")

	r.HandleFunc("/ranges", rangeHandler.ListRanges).Methods("GET")
	r.HandleFunc("/ranges", rangeHandler.AddRange).Methods("POST")
	r.HandleFunc("/ranges/{id}", rangeHandler.DeleteRange).Methods("DELETE")

	r.PathPrefix("/swagger/").Handler(httpSwagger.WrapHandler)

	corsOptions := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Content-Type"},
		AllowCredentials: true,
	})

	appLogger.Info("Server listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", corsOptions.Handler(r)))
}
