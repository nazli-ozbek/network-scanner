package repository

import (
	"database/sql"
	"encoding/json"
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
	if err := ensureDevicesTable(db); err != nil {
		logger.Error("failed to create devices table", err)
		return nil, err
	}
	if err := ensureIPRangesTable(db); err != nil {
		logger.Error("failed to create ip_ranges table", err)
		return nil, err
	}
	return &SQLiteRepository{db: db, logger: logger}, nil
}

func NewSQLiteRepositoryWithDB(db *sql.DB, logger logger.Logger) (*SQLiteRepository, error) {
	if err := ensureDevicesTable(db); err != nil {
		logger.Error("failed to create devices table", err)
		return nil, err
	}
	if err := ensureIPRangesTable(db); err != nil {
		logger.Error("failed to create ip_ranges table", err)
		return nil, err
	}
	return &SQLiteRepository{db: db, logger: logger}, nil
}

func ensureDevicesTable(db *sql.DB) error {
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS devices (
			id TEXT PRIMARY KEY,
			ip_address TEXT NOT NULL,
			mac_address TEXT,
			hostname TEXT,
			status TEXT,
			manufacturer TEXT,
			tags TEXT,
			last_seen DATETIME,
			first_seen DATETIME
		);
		CREATE UNIQUE INDEX IF NOT EXISTS idx_devices_ip ON devices(ip_address);
		CREATE INDEX IF NOT EXISTS idx_devices_hostname ON devices(hostname);
	`)
	return err
}

func ensureIPRangesTable(db *sql.DB) error {
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS ip_ranges (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			range TEXT NOT NULL
		);
	`)
	return err
}

func (r *SQLiteRepository) Save(d model.Device) {
	tagsJSON, _ := json.Marshal(d.Tags)
	existing := r.FindByIP(d.IPAddress)
	if existing != nil {
		if len(d.Tags) == 0 {
			d.Tags = existing.Tags
		}
		d.FirstSeen = existing.FirstSeen
	}
	_, err := r.db.Exec(`
		INSERT OR REPLACE INTO devices
			(id, ip_address, mac_address, hostname, status, manufacturer, tags, last_seen, first_seen)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`,
		d.ID,
		d.IPAddress,
		d.MACAddress,
		d.Hostname,
		d.Status,
		d.Manufacturer,
		string(tagsJSON),
		d.LastSeen.UTC().Format(time.RFC3339),
		d.FirstSeen.UTC().Format(time.RFC3339),
	)
	if err != nil {
		r.logger.Error("SQLite Save error", err)
	}
}

func (r *SQLiteRepository) GetAll() []model.Device {
	rows, err := r.db.Query(`
		SELECT id, ip_address, mac_address, hostname, status, manufacturer, tags, last_seen, first_seen
		FROM devices
	`)
	if err != nil {
		r.logger.Error("SQLite GetAll error", err)
		return nil
	}
	defer rows.Close()

	var out []model.Device
	for rows.Next() {
		var d model.Device
		var tagsRaw string
		var lastSeenStr, firstSeenStr string
		if err := rows.Scan(&d.ID, &d.IPAddress, &d.MACAddress, &d.Hostname, &d.Status, &d.Manufacturer, &tagsRaw, &lastSeenStr, &firstSeenStr); err == nil {
			_ = json.Unmarshal([]byte(defaultIfEmpty(tagsRaw, "[]")), &d.Tags)
			if lastSeenStr != "" {
				d.LastSeen, _ = time.Parse(time.RFC3339, lastSeenStr)
			}
			if firstSeenStr != "" {
				d.FirstSeen, _ = time.Parse(time.RFC3339, firstSeenStr)
			}
			out = append(out, d)
		}
	}
	return out
}

func (r *SQLiteRepository) Clear() {
	_, err := r.db.Exec("DELETE FROM devices")
	if err != nil {
		r.logger.Error("SQLite Clear error", err)
	}
}

func (r *SQLiteRepository) FindByID(id string) (*model.Device, error) {
	row := r.db.QueryRow(`
		SELECT id, ip_address, mac_address, hostname, status, manufacturer, tags, last_seen, first_seen
		FROM devices WHERE id = ?
	`, id)
	var d model.Device
	var tagsRaw string
	var lastSeenStr, firstSeenStr string
	if err := row.Scan(&d.ID, &d.IPAddress, &d.MACAddress, &d.Hostname, &d.Status, &d.Manufacturer, &tagsRaw, &lastSeenStr, &firstSeenStr); err != nil {
		return nil, err
	}
	_ = json.Unmarshal([]byte(defaultIfEmpty(tagsRaw, "[]")), &d.Tags)
	if lastSeenStr != "" {
		d.LastSeen, _ = time.Parse(time.RFC3339, lastSeenStr)
	}
	if firstSeenStr != "" {
		d.FirstSeen, _ = time.Parse(time.RFC3339, firstSeenStr)
	}
	return &d, nil
}

func (r *SQLiteRepository) UpdateTags(id string, tags []string) error {
	b, _ := json.Marshal(tags)
	_, err := r.db.Exec(`UPDATE devices SET tags = ? WHERE id = ?`, string(b), id)
	return err
}

func (r *SQLiteRepository) Search(q string) ([]model.Device, error) {
	like := "%" + q + "%"
	rows, err := r.db.Query(`
    SELECT id, ip_address, mac_address, hostname, status, manufacturer, tags, last_seen, first_seen
    FROM devices
    WHERE ip_address LIKE ? 
    OR mac_address LIKE ? 
    OR hostname LIKE ? 
    OR manufacturer LIKE ?
    OR tags LIKE ?
`, like, like, like, like, like)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []model.Device
	for rows.Next() {
		var d model.Device
		var tagsRaw string
		var lastSeenStr, firstSeenStr string
		if err := rows.Scan(&d.ID, &d.IPAddress, &d.MACAddress, &d.Hostname, &d.Status, &d.Manufacturer, &tagsRaw, &lastSeenStr, &firstSeenStr); err == nil {
			_ = json.Unmarshal([]byte(defaultIfEmpty(tagsRaw, "[]")), &d.Tags)
			if lastSeenStr != "" {
				d.LastSeen, _ = time.Parse(time.RFC3339, lastSeenStr)
			}
			if firstSeenStr != "" {
				d.FirstSeen, _ = time.Parse(time.RFC3339, firstSeenStr)
			}
			out = append(out, d)
		}
	}
	return out, nil
}

func (r *SQLiteRepository) FindByIP(ip string) *model.Device {
	row := r.db.QueryRow(`
		SELECT id, ip_address, mac_address, hostname, status, manufacturer, tags, last_seen, first_seen
		FROM devices WHERE ip_address = ?
	`, ip)
	var d model.Device
	var tagsRaw string
	var lastSeenStr, firstSeenStr string
	err := row.Scan(&d.ID, &d.IPAddress, &d.MACAddress, &d.Hostname, &d.Status, &d.Manufacturer, &tagsRaw, &lastSeenStr, &firstSeenStr)
	if err != nil {
		if err == sql.ErrNoRows {
			r.logger.Debug("Device not found in DB:", ip)
			return nil
		}
		r.logger.Error("SQLite FindByIP error:", err)
		return nil
	}
	_ = json.Unmarshal([]byte(defaultIfEmpty(tagsRaw, "[]")), &d.Tags)
	if lastSeenStr != "" {
		d.LastSeen, _ = time.Parse(time.RFC3339, lastSeenStr)
	}
	if firstSeenStr != "" {
		d.FirstSeen, _ = time.Parse(time.RFC3339, firstSeenStr)
	}
	return &d
}

func defaultIfEmpty(s, def string) string {
	if s == "" {
		return def
	}
	return s
}

var _ DeviceRepository = (*SQLiteRepository)(nil)
