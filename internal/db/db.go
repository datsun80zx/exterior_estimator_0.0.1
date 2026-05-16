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
		`CREATE TABLE IF NOT EXISTS materials (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			category TEXT NOT NULL CHECK(category IN ('roofing', 'siding', 'gutters')),
			brand TEXT NOT NULL DEFAULT '',
			product_line TEXT NOT NULL DEFAULT '',
			name TEXT NOT NULL,
			unit TEXT NOT NULL,
			coverage_per_unit REAL NOT NULL DEFAULT 0,
			cost_per_unit REAL NOT NULL DEFAULT 0,
			price_per_unit REAL NOT NULL DEFAULT 0,
			is_active INTEGER NOT NULL DEFAULT 1,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS labor_rates (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			category TEXT NOT NULL CHECK(category IN ('roofing', 'siding', 'gutters')),
			description TEXT NOT NULL,
			rate_per_sq REAL NOT NULL DEFAULT 0,
			rate_per_lf REAL NOT NULL DEFAULT 0,
			rate_per_sqft REAL NOT NULL DEFAULT 0,
			is_active INTEGER NOT NULL DEFAULT 1
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

// --- Material Queries ---

func (s *Store) GetMaterialsByCategory(category models.MaterialCategory) ([]models.Material, error) {
	rows, err := s.DB.Query(`
		SELECT id, category, brand, product_line, name, unit,
		       coverage_per_unit, cost_per_unit, price_per_unit, is_active
		FROM materials
		WHERE category = ? AND is_active = 1
		ORDER BY brand, product_line, name
	`, string(category))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var materials []models.Material
	for rows.Next() {
		var m models.Material
		if err := rows.Scan(&m.ID, &m.Category, &m.Brand, &m.ProductLine, &m.Name, &m.Unit,
			&m.CoveragePerUnit, &m.CostPerUnit, &m.PricePerUnit, &m.IsActive); err != nil {
			return nil, err
		}
		materials = append(materials, m)
	}
	return materials, nil
}

func (s *Store) GetAllMaterials() ([]models.Material, error) {
	rows, err := s.DB.Query(`
		SELECT id, category, brand, product_line, name, unit,
		       coverage_per_unit, cost_per_unit, price_per_unit, is_active
		FROM materials
		ORDER BY category, brand, product_line, name
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var materials []models.Material
	for rows.Next() {
		var m models.Material
		if err := rows.Scan(&m.ID, &m.Category, &m.Brand, &m.ProductLine, &m.Name, &m.Unit,
			&m.CoveragePerUnit, &m.CostPerUnit, &m.PricePerUnit, &m.IsActive); err != nil {
			return nil, err
		}
		materials = append(materials, m)
	}
	return materials, nil
}

func (s *Store) GetMaterial(id int64) (models.Material, error) {
	var m models.Material
	err := s.DB.QueryRow(`
		SELECT id, category, brand, product_line, name, unit,
		       coverage_per_unit, cost_per_unit, price_per_unit, is_active
		FROM materials WHERE id = ?
	`, id).Scan(&m.ID, &m.Category, &m.Brand, &m.ProductLine, &m.Name, &m.Unit,
		&m.CoveragePerUnit, &m.CostPerUnit, &m.PricePerUnit, &m.IsActive)
	return m, err
}

func (s *Store) UpsertMaterial(m models.Material) (int64, error) {
	if m.ID > 0 {
		_, err := s.DB.Exec(`
			UPDATE materials SET
				category = ?, brand = ?, product_line = ?, name = ?, unit = ?,
				coverage_per_unit = ?, cost_per_unit = ?, price_per_unit = ?,
				is_active = ?, updated_at = CURRENT_TIMESTAMP
			WHERE id = ?
		`, m.Category, m.Brand, m.ProductLine, m.Name, m.Unit,
			m.CoveragePerUnit, m.CostPerUnit, m.PricePerUnit, m.IsActive, m.ID)
		return m.ID, err
	}

	res, err := s.DB.Exec(`
		INSERT INTO materials (category, brand, product_line, name, unit,
		                       coverage_per_unit, cost_per_unit, price_per_unit, is_active)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, m.Category, m.Brand, m.ProductLine, m.Name, m.Unit,
		m.CoveragePerUnit, m.CostPerUnit, m.PricePerUnit, m.IsActive)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (s *Store) DeleteMaterial(id int64) error {
	_, err := s.DB.Exec("UPDATE materials SET is_active = 0, updated_at = CURRENT_TIMESTAMP WHERE id = ?", id)
	return err
}

// --- Shingle Brands (convenience) ---

func (s *Store) GetShingleBrands() ([]string, error) {
	rows, err := s.DB.Query(`
		SELECT DISTINCT brand FROM materials
		WHERE category = 'roofing' AND name LIKE '%Shingle%' AND is_active = 1
		ORDER BY brand
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var brands []string
	for rows.Next() {
		var b string
		if err := rows.Scan(&b); err != nil {
			return nil, err
		}
		brands = append(brands, b)
	}
	return brands, nil
}

func (s *Store) GetShinglesByBrand(brand string) ([]models.Material, error) {
	rows, err := s.DB.Query(`
		SELECT id, category, brand, product_line, name, unit,
		       coverage_per_unit, cost_per_unit, price_per_unit, is_active
		FROM materials
		WHERE category = 'roofing' AND brand = ? AND name LIKE '%Shingle%' AND is_active = 1
		ORDER BY product_line, name
	`, brand)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var materials []models.Material
	for rows.Next() {
		var m models.Material
		if err := rows.Scan(&m.ID, &m.Category, &m.Brand, &m.ProductLine, &m.Name, &m.Unit,
			&m.CoveragePerUnit, &m.CostPerUnit, &m.PricePerUnit, &m.IsActive); err != nil {
			return nil, err
		}
		materials = append(materials, m)
	}
	return materials, nil
}

// --- Labor Rates ---

func (s *Store) GetLaborRate(category models.MaterialCategory) (models.LaborRate, error) {
	var lr models.LaborRate
	err := s.DB.QueryRow(`
		SELECT id, category, description, rate_per_sq, rate_per_lf, rate_per_sqft
		FROM labor_rates WHERE category = ? AND is_active = 1
	`, string(category)).Scan(&lr.ID, &lr.Category, &lr.Description,
		&lr.RatePerSq, &lr.RatePerLF, &lr.RatePerSqFt)
	if err == sql.ErrNoRows {
		return models.LaborRate{Category: category}, nil
	}
	return lr, err
}

func (s *Store) UpsertLaborRate(lr models.LaborRate) error {
	if lr.ID > 0 {
		_, err := s.DB.Exec(`
			UPDATE labor_rates SET description = ?, rate_per_sq = ?, rate_per_lf = ?, rate_per_sqft = ?
			WHERE id = ?
		`, lr.Description, lr.RatePerSq, lr.RatePerLF, lr.RatePerSqFt, lr.ID)
		return err
	}
	_, err := s.DB.Exec(`
		INSERT INTO labor_rates (category, description, rate_per_sq, rate_per_lf, rate_per_sqft)
		VALUES (?, ?, ?, ?, ?)
	`, lr.Category, lr.Description, lr.RatePerSq, lr.RatePerLF, lr.RatePerSqFt)
	return err
}

// --- Saved Estimates ---

func (s *Store) SaveEstimate(customerName, customerAddress, projectType, estimateJSON string) (int64, error) {
	res, err := s.DB.Exec(`
		INSERT INTO saved_estimates (customer_name, customer_address, project_type, estimate_json)
		VALUES (?, ?, ?, ?)
	`, customerName, customerAddress, projectType, estimateJSON)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

// Seed checks if materials table is empty and seeds if so
func (s *Store) SeedIfEmpty() error {
	var count int
	if err := s.DB.QueryRow("SELECT COUNT(*) FROM materials").Scan(&count); err != nil {
		return err
	}
	if count > 0 {
		log.Println("Database already seeded, skipping")
		return nil
	}
	return s.seed()
}
