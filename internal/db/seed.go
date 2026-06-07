package db

import "log"

func (s *Store) seed() error {
	log.Println("Seeding database with default materials...")

	materials := []struct {
		category, brand, productLine, name, unit   string
		coveragePerUnit, costPerUnit, pricePerUnit float64
	}{
		// =============================================
		// SHINGLES - Owens Corning
		// =============================================
		// Coverage: 33.3 sq ft per bundle (3 bundles = 1 square)
		{"roofing", "Owens Corning", "Duration", "Duration Shingles", "bundle", 33.3, 0, 0},
		{"roofing", "Owens Corning", "Duration STORM", "Duration STORM Shingles", "bundle", 33.3, 0, 0},
		{"roofing", "Owens Corning", "Duration FLEX", "Duration FLEX Shingles", "bundle", 33.3, 0, 0},
		{"roofing", "Owens Corning", "TruDefinition Duration", "TruDefinition Duration Shingles", "bundle", 33.3, 0, 0},
		{"roofing", "Owens Corning", "Oakridge", "Oakridge Shingles", "bundle", 33.3, 0, 0},

		// =============================================
		// SHINGLES - GAF
		// =============================================
		{"roofing", "GAF", "Timberline HDZ", "Timberline HDZ Shingles", "bundle", 33.3, 0, 0},
		{"roofing", "GAF", "Timberline AS II", "Timberline AS II Shingles", "bundle", 33.3, 0, 0},
		{"roofing", "GAF", "Timberline NS", "Timberline NS Shingles", "bundle", 33.3, 0, 0},
		{"roofing", "GAF", "Royal Sovereign", "Royal Sovereign 3-Tab Shingles", "bundle", 33.3, 0, 0},

		// =============================================
		// SHINGLES - CertainTeed
		// =============================================
		{"roofing", "CertainTeed", "Landmark", "Landmark Shingles", "bundle", 33.3, 0, 0},
		{"roofing", "CertainTeed", "Landmark PRO", "Landmark PRO Shingles", "bundle", 33.3, 0, 0},
		{"roofing", "CertainTeed", "NorthGate", "NorthGate SBS Shingles", "bundle", 33.3, 0, 0},

		// =============================================
		// UNDERLAYMENT
		// =============================================
		// Synthetic underlayment - typical roll covers 1000 sq ft
		{"roofing", "Owens Corning", "ProArmor", "ProArmor Synthetic Underlayment", "roll", 1000, 0, 0},
		{"roofing", "GAF", "FeltBuster", "FeltBuster Synthetic Underlayment", "roll", 1000, 0, 0},
		{"roofing", "Generic", "", "15# Felt Underlayment", "roll", 400, 0, 0},
		{"roofing", "Generic", "", "30# Felt Underlayment", "roll", 200, 0, 0},

		// =============================================
		// ICE & WATER SHIELD
		// =============================================
		// Typical roll: 2 sq (200 sq ft) - 36" x 66.7'
		{"roofing", "Owens Corning", "WeatherLock", "WeatherLock G Ice & Water", "roll", 200, 0, 0},
		{"roofing", "GAF", "StormGuard", "StormGuard Ice & Water", "roll", 200, 0, 0},
		{"roofing", "Grace", "Ice & Water Shield", "Grace Ice & Water Shield", "roll", 200, 0, 0},

		// =============================================
		// DRIP EDGE
		// =============================================
		// 10 ft pieces
		{"roofing", "Generic", "", "Drip Edge - Aluminum (White)", "piece", 10, 0, 0},
		{"roofing", "Generic", "", "Drip Edge - Aluminum (Brown)", "piece", 10, 0, 0},
		{"roofing", "Generic", "", "Drip Edge - Aluminum (Black)", "piece", 10, 0, 0},

		// =============================================
		// RIDGE CAP
		// =============================================
		// OC hip & ridge covers ~31.7 linear ft per bundle
		{"roofing", "Owens Corning", "DecoRidge", "DecoRidge Hip & Ridge", "bundle", 20, 0, 0},
		{"roofing", "Owens Corning", "", "ProEdge Hip & Ridge", "bundle", 31, 0, 0},
		{"roofing", "GAF", "TimberTex", "TimberTex Hip & Ridge", "bundle", 20, 0, 0},
		{"roofing", "GAF", "Seal-A-Ridge", "Seal-A-Ridge Hip & Ridge", "bundle", 25, 0, 0},
		{"roofing", "CertainTeed", "", "Shadow Ridge Hip & Ridge", "bundle", 31.7, 0, 0},

		// =============================================
		// STARTER STRIP
		// =============================================
		// Starter strip ~ 105 linear ft per bundle (varies)
		{"roofing", "Owens Corning", "Starter Strip Plus", "Starter Strip Shingles", "bundle", 105, 0, 0},
		{"roofing", "GAF", "Pro-Start", "Pro-Start Starter Strip", "bundle", 120, 0, 0},
		{"roofing", "CertainTeed", "", "SwiftStart Starter Shingles", "bundle", 117, 0, 0},

		// =============================================
		// PIPE BOOTS & VENTS
		// =============================================
		{"roofing", "Generic", "", "Pipe Boot - 1.5\" to 3\"", "piece", 1, 0, 0},
		{"roofing", "Generic", "", "Pipe Boot - 3\" to 4\"", "piece", 1, 0, 0},
		{"roofing", "Generic", "", "Exhaust Vent (box vent)", "piece", 1, 0, 0},
		{"roofing", "Owens Corning", "", "VentSure Ridge Vent (4ft)", "piece", 4, 0, 0},
		{"roofing", "Owens Corning", "Ridge Prowler 30", "Ridge Prowler", "roll", 30, 0, 0},

		// =============================================
		// NAILS & MISC 1 box of nails covers 15 SQ
		// =============================================
		{"roofing", "Generic", "", "1-1/4\" Coil Roofing Nails", "box", 15, 0, 0},
		{"roofing", "Generic", "", "Roofing Caulk / Sealant", "tube", 1, 0, 0},

		// =============================================
		// LABOR RATES (seeded separately below)
		// =============================================
	}

	tx, err := s.DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(`
		INSERT INTO materials (category, brand, product_line, name, unit,
		                       coverage_per_unit, cost_per_unit, price_per_unit, is_active)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, 1)
	`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, m := range materials {
		if _, err := stmt.Exec(m.category, m.brand, m.productLine, m.name, m.unit,
			m.coveragePerUnit, m.costPerUnit, m.pricePerUnit); err != nil {
			return err
		}
	}

	// Seed default labor rates (placeholder - you'll set these)
	laborRates := []struct {
		category, description             string
		ratePerSq, ratePerLF, ratePerSqFt float64
	}{
		{"roofing", "install", 300, 0, 0}, // $300/square labor
		{"roofing", "Tear-off (per layer per square)", 90, 0, 0},
		{"gutters", "Standard gutter install", 0, 6, 0},   // $6/LF
		{"siding", "Standard siding install", 0, 0, 3.50}, // $3.50/sqft
	}

	lrStmt, err := tx.Prepare(`
		INSERT INTO labor_rates (category, description, rate_per_sq, rate_per_lf, rate_per_sqft)
		VALUES (?, ?, ?, ?, ?)
	`)
	if err != nil {
		return err
	}
	defer lrStmt.Close()

	for _, lr := range laborRates {
		if _, err := lrStmt.Exec(lr.category, lr.description, lr.ratePerSq, lr.ratePerLF, lr.ratePerSqFt); err != nil {
			return err
		}
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	log.Println("Database seeded successfully")
	return nil
}
