package db

import (
	"log"

	"github.com/datsun80zx/exterior_estimator_0.0.1/internal/models"
)

// seed loads the price book with the exact rates from the May Customer Price
// Book (Master Copy, column B).
func (s *Store) seed() error {
	log.Println("Seeding price book with rates from the Customer Price Book...")

	type seedRate struct {
		key, label string
		group      models.RateGroup
		unit       string
		amount     float64
	}

	rates := []seedRate{
		// Shingles ($/sqft)
		{models.KeyShingleOakridge, "OC Oakridge Architectural", models.GroupShingle, "/sqft", 1.21},
		{models.KeyShingleDuration, "OC TruDefinition Duration", models.GroupShingle, "/sqft", 1.40},
		{models.KeyShingleDesigner, "OC TruDefinition Duration Designer", models.GroupShingle, "/sqft", 1.40},
		{models.KeyShingleFlex, "OC TruDefinition Duration Flex (Class 4)", models.GroupShingle, "/sqft", 1.60},
		{models.KeyShingleBrava, "Brava Faux Slate (15 bundles/sq)", models.GroupShingle, "/sqft", 8.05},
		{models.KeyShingleGrandManor, "CertainTeed Grand Manor Lux", models.GroupShingle, "/sqft", 3.95},

		// Hip & Ridge ($/lf)
		{models.KeyHRProEdge, "OC ProEdge Hip & Ridge", models.GroupHipRidge, "/lf", 2.40},
		{models.KeyHRProEdgeFlex, "OC ProEdge Hip & Ridge Flex", models.GroupHipRidge, "/lf", 3.06},
		{models.KeyHRBrava, "Brava Faux Slate Hip & Ridge", models.GroupHipRidge, "/lf", 9.04},
		{models.KeyHRGrandManor, "CertainTeed Grand Manor Hip & Ridge", models.GroupHipRidge, "/lf", 4.85},

		// Starter ($/lf)
		{models.KeyStarter, "Starter Strip Plus", models.GroupStarter, "/lf", 0.65},

		// Ice & Water ($/lf)
		{models.KeyIWWeatherlockG, "Weatherlock G", models.GroupIceWater, "/lf", 1.60},
		{models.KeyIWWeatherlockMat, "Weatherlock Mat", models.GroupIceWater, "/lf", 1.60},
		{models.KeyIWWeatherlockFlex, "WeatherLock Flex", models.GroupIceWater, "/lf", 2.57},

		// Underlayment ($/sqft)
		{models.KeyUnderProArmor, "OC ProArmor Synthetic Underlayment", models.GroupUnderlay, "/sqft", 0.15},
		{models.KeyUnderDeckDefense, "OC DeckDefense", models.GroupUnderlay, "/sqft", 0.20},
		{models.KeyUnderRhinoRoof, "OC RhinoRoof U20", models.GroupUnderlay, "/sqft", 0.08},

		// Misc
		{models.KeyPipeBoot, "Pipe Boot", models.GroupMisc, "/ea", 11.60},
		{models.KeyNails, "Nails", models.GroupMisc, "/sqft", 0.04},
		{models.KeyFlashing, "Flashing", models.GroupMisc, "/ea", 100},
		{models.KeyDripEdge, "Drip Edge", models.GroupMisc, "/lf", 0.80},
		{models.KeyRidgeVent, "Ridge Vent", models.GroupMisc, "/lf", 4.75},
		{models.KeyOSB, "OSB", models.GroupMisc, "/sheet", 85},
		{models.KeyWarranty, "Warranty", models.GroupMisc, "/sqft", 0.14},
		{models.KeyPlankDeck, "Plank Deck", models.GroupMisc, "/lf", 8.85},
		{models.KeySolarFan, "Solar Fan", models.GroupMisc, "/ea", 542.99},
		{models.KeyPowerFan, "Power Fan", models.GroupMisc, "/ea", 187.56},
		{models.KeyBroanVent, "Broan Vent", models.GroupMisc, "/ea", 48.53},
		{models.KeyDumpster, "Dumpster (per 15 sq)", models.GroupMisc, "/ea", 500},
		{models.KeySkylight, "Skylight (flat default)", models.GroupMisc, "/ea", 1400},

		// Labor ($/sqft)
		{models.KeyLabor, "Labor", models.GroupLabor, "/sqft", 3.00},
		{models.KeyLaborSpecialty, "Labor (Specialty)", models.GroupLabor, "/sqft", 4.50},
		{models.KeyAddtlTearoff, "Additional Tear-off", models.GroupLabor, "/sqft", 0.90},
		{models.KeySteepSlope, "Steep Slope (>8/12)", models.GroupLabor, "/sqft", 0.25},

		// Settings
		{models.KeyBusinessMargin, "Business Margin (price divisor)", models.GroupSetting, "factor", 0.85},
	}

	tx, err := s.DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(`INSERT INTO rates (key, label, grp, unit, amount) VALUES (?, ?, ?, ?, ?)`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, r := range rates {
		if _, err := stmt.Exec(r.key, r.label, string(r.group), r.unit, r.amount); err != nil {
			return err
		}
	}

	if err := tx.Commit(); err != nil {
		return err
	}
	log.Println("Price book seeded successfully")
	return nil
}
