package db

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/mattn/go-sqlite3"

	"github.com/datsun80zx/exterior_estimator_0.0.1/internal/models"
)

type Store struct {
	DB *sql.DB
}

func New(dbPath string) (*Store, error) {
	db, err := sql.Open("sqlite3", dbPath+"?_journal_mode=WAL&_foreign_keys=on")
	if err != nil {
		return nil, fmt.Errorf("opening database: %w", err)
	}
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("pinging database: %w", err)
	}

	store := &Store{DB: db}
	if err := store.migrate(); err != nil {
		return nil, fmt.Errorf("running migrations: %w", err)
	}
	return store, nil
}

func (s *Store) migrate() error {
	migrations := []string{
		`CREATE TABLE IF NOT EXISTS rates (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			key TEXT NOT NULL UNIQUE,
			label TEXT NOT NULL,
			grp TEXT NOT NULL,
			unit TEXT NOT NULL DEFAULT '',
			amount REAL NOT NULL DEFAULT 0,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS saved_estimates (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			customer_name TEXT NOT NULL,
			customer_address TEXT NOT NULL DEFAULT '',
			project_type TEXT NOT NULL,
			estimate_json TEXT NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
	}
	for _, m := range migrations {
		if _, err := s.DB.Exec(m); err != nil {
			return fmt.Errorf("migration: %w", err)
		}
	}
	return nil
}

// --- Rate queries ---

func (s *Store) GetRates() ([]models.Rate, error) {
	rows, err := s.DB.Query(`SELECT id, key, label, grp, unit, amount FROM rates ORDER BY grp, label`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []models.Rate
	for rows.Next() {
		var r models.Rate
		if err := rows.Scan(&r.ID, &r.Key, &r.Label, &r.Group, &r.Unit, &r.Amount); err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	return out, rows.Err()
}

// GetRatesByGroup returns rates bucketed by group, in a stable group order.
func (s *Store) GetRatesByGroup() ([]struct {
	Group string
	Rates []models.Rate
}, error) {
	all, err := s.GetRates()
	if err != nil {
		return nil, err
	}
	order := []models.RateGroup{
		models.GroupShingle, models.GroupHipRidge, models.GroupStarter,
		models.GroupIceWater, models.GroupUnderlay, models.GroupMisc,
		models.GroupLabor, models.GroupSetting,
	}
	byGroup := map[models.RateGroup][]models.Rate{}
	for _, r := range all {
		byGroup[r.Group] = append(byGroup[r.Group], r)
	}
	var out []struct {
		Group string
		Rates []models.Rate
	}
	for _, g := range order {
		if rs := byGroup[g]; len(rs) > 0 {
			out = append(out, struct {
				Group string
				Rates []models.Rate
			}{Group: string(g), Rates: rs})
		}
	}
	return out, nil
}

// RateMap returns key -> amount for the calc engine.
func (s *Store) RateMap() (map[string]float64, error) {
	rows, err := s.DB.Query(`SELECT key, amount FROM rates`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	m := map[string]float64{}
	for rows.Next() {
		var k string
		var a float64
		if err := rows.Scan(&k, &a); err != nil {
			return nil, err
		}
		m[k] = a
	}
	return m, rows.Err()
}

// RateLabel returns the display label for a rate key.
func (s *Store) RateLabel(key string) (string, bool) {
	var lbl string
	err := s.DB.QueryRow(`SELECT label FROM rates WHERE key = ?`, key).Scan(&lbl)
	if err != nil {
		return "", false
	}
	return lbl, true
}

func (s *Store) GetRate(id int64) (models.Rate, error) {
	var r models.Rate
	err := s.DB.QueryRow(`SELECT id, key, label, grp, unit, amount FROM rates WHERE id = ?`, id).
		Scan(&r.ID, &r.Key, &r.Label, &r.Group, &r.Unit, &r.Amount)
	return r, err
}

// UpdateRateAmount updates just the dollar amount for an existing rate.
func (s *Store) UpdateRateAmount(id int64, amount float64) error {
	_, err := s.DB.Exec(`UPDATE rates SET amount = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?`, amount, id)
	return err
}

// --- Saved estimates ---

func (s *Store) SaveEstimate(customerName, customerAddress, projectType, estimateJSON string) (int64, error) {
	res, err := s.DB.Exec(`
		INSERT INTO saved_estimates (customer_name, customer_address, project_type, estimate_json)
		VALUES (?, ?, ?, ?)`,
		customerName, customerAddress, projectType, estimateJSON)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (s *Store) GetEstimateJSON(id string) (string, error) {
	var j string
	err := s.DB.QueryRow(`SELECT estimate_json FROM saved_estimates WHERE id = ?`, id).Scan(&j)
	return j, err
}

// --- Seed gate ---

func (s *Store) SeedIfEmpty() error {
	var count int
	if err := s.DB.QueryRow(`SELECT COUNT(*) FROM rates`).Scan(&count); err != nil {
		return err
	}
	if count > 0 {
		log.Println("Price book already seeded, skipping")
		return nil
	}
	return s.seed()
}
