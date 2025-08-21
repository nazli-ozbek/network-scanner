package repository

import (
	"database/sql"
	"network-scanner/logger"
	"network-scanner/model"

	_ "github.com/mattn/go-sqlite3"
)

type SQLiteIPRangeRepository struct {
	db     *sql.DB
	logger logger.Logger
}

func NewSQLiteIPRangeRepository(db *sql.DB, logger logger.Logger) *SQLiteIPRangeRepository {
	_, _ = db.Exec(`
		CREATE TABLE IF NOT EXISTS ip_ranges (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			range TEXT NOT NULL
		);
	`)
	return &SQLiteIPRangeRepository{db: db, logger: logger}
}

func (r *SQLiteIPRangeRepository) Save(x model.IPRange) error {
	_, err := r.db.Exec(`INSERT OR REPLACE INTO ip_ranges(id,name,range) VALUES (?,?,?)`, x.ID, x.Name, x.Range)
	return err
}

func (r *SQLiteIPRangeRepository) GetAll() ([]model.IPRange, error) {
	rows, err := r.db.Query(`SELECT id,name,range FROM ip_ranges ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []model.IPRange
	for rows.Next() {
		var x model.IPRange
		if err := rows.Scan(&x.ID, &x.Name, &x.Range); err == nil {
			out = append(out, x)
		}
	}
	return out, nil
}

func (r *SQLiteIPRangeRepository) Delete(id string) error {
	_, err := r.db.Exec(`DELETE FROM ip_ranges WHERE id=?`, id)
	return err
}
