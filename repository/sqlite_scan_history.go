package repository

import (
	"database/sql"
	"network-scanner/logger"
	"network-scanner/model"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type SQLiteScanHistoryRepository struct {
	db     *sql.DB
	logger logger.Logger
}

func NewSQLiteScanHistoryRepository(db *sql.DB, logger logger.Logger) (*SQLiteScanHistoryRepository, error) {
	createTable := `
	CREATE TABLE IF NOT EXISTS scan_history (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		ip_range TEXT NOT NULL,
		started_at DATETIME NOT NULL,
		device_count INTEGER NOT NULL
	);`
	if _, err := db.Exec(createTable); err != nil {
		logger.Error("failed to create scan_history table", err)
		return nil, err
	}
	return &SQLiteScanHistoryRepository{db: db, logger: logger}, nil
}

func (r *SQLiteScanHistoryRepository) Save(h model.ScanHistory) (int64, error) {
	stmt := `INSERT INTO scan_history (ip_range, started_at, device_count) VALUES (?, ?, ?)`
	res, err := r.db.Exec(stmt, h.IPRange, h.StartedAt.UTC().Format(time.RFC3339), h.DeviceCount)
	if err != nil {
		r.logger.Error("SQLite scan_history Save error", err)
		return 0, err
	}
	return res.LastInsertId()
}

func (r *SQLiteScanHistoryRepository) GetAll() ([]model.ScanHistory, error) {
	rows, err := r.db.Query(`SELECT id, ip_range, started_at, device_count FROM scan_history ORDER BY id DESC`)
	if err != nil {
		r.logger.Error("SQLite scan_history GetAll error", err)
		return nil, err
	}
	defer rows.Close()

	var out []model.ScanHistory
	for rows.Next() {
		var h model.ScanHistory
		var started string
		if err := rows.Scan(&h.ID, &h.IPRange, &started, &h.DeviceCount); err == nil {
			h.StartedAt, _ = time.Parse(time.RFC3339, started)
			out = append(out, h)
		}
	}
	return out, nil
}

func (r *SQLiteScanHistoryRepository) GetByID(id int64) (*model.ScanHistory, error) {
	row := r.db.QueryRow(`SELECT id, ip_range, started_at, device_count FROM scan_history WHERE id = ?`, id)
	var h model.ScanHistory
	var started string
	if err := row.Scan(&h.ID, &h.IPRange, &started, &h.DeviceCount); err != nil {
		return nil, err
	}
	h.StartedAt, _ = time.Parse(time.RFC3339, started)
	return &h, nil
}

func (r *SQLiteScanHistoryRepository) Delete(id int64) error {
	_, err := r.db.Exec("DELETE FROM scan_history WHERE id = ?", id)
	if err != nil {
		r.logger.Error("SQLite scan_history Delete error", err)
		return err
	}
	return nil
}

func (r *SQLiteScanHistoryRepository) Clear() error {
	_, err := r.db.Exec("DELETE FROM scan_history")
	if err != nil {
		r.logger.Error("SQLite scan_history Clear error", err)
		return err
	}
	return nil
}
