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

	"network-scanner/api"
	"network-scanner/config"
	_ "network-scanner/docs"
	"network-scanner/logger"
	"network-scanner/repository"
	"network-scanner/service"

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

	historyRepo, err := repository.NewSQLiteScanHistoryRepository(db, appLogger)
	if err != nil {
		appLogger.Error("Failed to initialize scan history repo:", err)
		return
	}

	scanner := service.NewScannerService(deviceRepo, historyRepo, appLogger)

	scanHandler := api.NewScanHandler(scanner, appLogger, historyRepo)
	deviceHandler := api.NewDeviceHandler(scanner, appLogger)
	histHandler := api.NewScanHistoryHandler(historyRepo, appLogger)

	mux := http.NewServeMux()
	mux.HandleFunc("/scan", scanHandler.StartScan)
	mux.HandleFunc("/scan/repeat", scanHandler.RepeatScan)
	mux.HandleFunc("/devices", deviceHandler.GetDevices)
	mux.HandleFunc("/clear", deviceHandler.ClearDevices)

	mux.HandleFunc("/scan-history", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			histHandler.GetScanHistory(w, r)
		case http.MethodDelete:
			histHandler.ClearScanHistory(w, r)
		default:
			http.NotFound(w, r)
		}
	})
	mux.HandleFunc("/scan-history/", histHandler.DeleteScanHistory)

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
