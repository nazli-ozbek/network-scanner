package repository

import (
	"database/sql"
	"network-scanner/logger"
	"network-scanner/model"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type SQLiteRepository struct {
	db     *sql.DB
	logger logger.Logger
}

func NewSQLiteRepository(dbPath string, logger logger.Logger) (*SQLiteRepository, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		logger.Error("failed to open sqlite db", err)
		return nil, err
	}

	createTable := `
	CREATE TABLE IF NOT EXISTS devices (
		ip_address TEXT PRIMARY KEY,
		mac_address TEXT,
		hostname TEXT,
		is_online BOOLEAN,
		last_seen DATETIME
	);
	`
	_, err = db.Exec(createTable)
	if err != nil {
		logger.Error("failed to create table", err)
		return nil, err
	}

	return &SQLiteRepository{db: db, logger: logger}, nil
}

func (r *SQLiteRepository) Save(device model.Device) {
	stmt := `
    INSERT OR REPLACE INTO devices(ip_address, mac_address, hostname, is_online, last_seen)
    VALUES (?, ?, ?, ?, ?)
    `
	_, err := r.db.Exec(stmt, device.IPAddress, device.MACAddress, device.Hostname, device.IsOnline, device.LastSeen)
	if err != nil {
		r.logger.Error("SQLite Save error", err)
	}

}

func (r *SQLiteRepository) GetAll() []model.Device {
	rows, err := r.db.Query("SELECT ip_address, mac_address, hostname, is_online, last_seen FROM devices")
	if err != nil {
		r.logger.Error("SQLite GetAll error", err)
		return nil
	}
	defer rows.Close()

	var devices []model.Device
	for rows.Next() {
		var d model.Device
		var lastSeenStr string
		err := rows.Scan(&d.IPAddress, &d.MACAddress, &d.Hostname, &d.IsOnline, &lastSeenStr)
		if err == nil {
			d.LastSeen, _ = time.Parse(time.RFC3339, lastSeenStr)
			devices = append(devices, d)
		}
	}
	return devices
}

func (r *SQLiteRepository) Clear() {
	_, err := r.db.Exec("DELETE FROM devices")
	if err != nil {
		r.logger.Error("SQLite Clear error", err)
	}
}

func (r *SQLiteRepository) FindByIP(ip string) *model.Device {
	row := r.db.QueryRow("SELECT ip_address, mac_address, hostname, is_online, last_seen FROM devices WHERE ip_address = ?", ip)
	var d model.Device
	var lastSeenStr string
	err := row.Scan(&d.IPAddress, &d.MACAddress, &d.Hostname, &d.IsOnline, &lastSeenStr)
	if err != nil {
		if err == sql.ErrNoRows {
			r.logger.Debug("Device not found in DB:", ip)
			return nil
		}
		r.logger.Error("SQLite FindByIP error:", err)
	}

	d.LastSeen, _ = time.Parse(time.RFC3339, lastSeenStr)
	return &d
}

var _ DeviceRepository = (*SQLiteRepository)(nil)
