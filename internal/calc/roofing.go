package calc

import (
	"math"
	"time"

	"github.com/datsun80zx/exterior_estimator_0.0.1/internal/db"
	"github.com/datsun80zx/exterior_estimator_0.0.1/internal/models"
)

type RoofingCalculator struct {
	store *db.Store
}

func NewRoofingCalculator(store *db.Store) *RoofingCalculator {
	return &RoofingCalculator{store: store}
}

// Calculate prices out all four tiers for one set of measurements, plus the
// shared supplier order list. The formulas mirror the validated Excel model:
//
//	Materials (per tier, summed):
//	  Shingles      = roof_sqft        * shingle_rate
//	  Dumpster      = dumpsters         * dumpster_rate
//	  Plywood       = decking_sheets    * osb_rate
//	  Planks        = plank_deck_lf     * plank_deck_rate
//	  Warranty      = roof_sqft         * warranty_rate
//	  Underlayment  = roof_sqft         * deckdefense_rate
//	  Ice & Water   = (eaves*2 + valleys + intake) * iw_rate
//	  Ridge Vent    = ridge_lf          * ridge_vent_rate
//	  Hip & Ridge   = (ridge_lf + hips) * hip_ridge_rate
//	  Intake Vent   = intake_lf         * ridge_vent_rate
//	  Starter       = (eaves + rakes)   * starter_rate
//	  Drip Edge     = (eaves + rakes)   * drip_edge_rate
//	  Nails         = roof_sqft         * nails_rate
//	  Solar Fan     = solar_fans        * solar_fan_rate
//	  Power Fan     = power_fans        * power_fan_rate
//	  Broan Vent    = broan_vents       * broan_vent_rate
//	  Pipe Boot     = pipe_boots        * pipeboot_rate
//	  Chimney Flash = chimney           * flashing_rate * 2.5
//	  Flashing      = flashing_rate (flat line, always included)
//	  Skylights     = skylight_total (matrix or flat)
//
//	Labor = roof_sqft * labor_rate + addtl_tearoff_sf * addtl_tearoff_rate
//	                                + steep_slope_sf  * steep_slope_rate
//	Subtotal   = Materials + Labor
//	TotalPrice = Subtotal / business_margin
//	Overhead   = TotalPrice - Subtotal
func (c *RoofingCalculator) Calculate(m models.RoofMeasurements) (*models.EstimateResult, error) {
	r, err := c.store.RateMap()
	if err != nil {
		return nil, err
	}

	margin := r[models.KeyBusinessMargin]
	if margin <= 0 {
		margin = 0.85
	}

	// Skylight total: detailed matrix selection if chosen, else flat rate.
	skylightTotal := float64(m.Skylights) * r[models.KeySkylight]
	if price, ok := models.LookupSkylight(m.SkylightMount, m.SkylightType, m.SkylightSize); ok {
		skylightTotal = float64(m.Skylights) * price
	}

	// Add-on / shared material lines (identical across all four tiers).
	shared := []models.LineItem{
		{Name: "Dumpster", Amount: float64(m.Dumpsters) * r[models.KeyDumpster]},
		{Name: "Plywood (OSB)", Amount: float64(m.Decking) * r[models.KeyOSB]},
		{Name: "Plank Decking", Amount: m.PlankDeckLF * r[models.KeyPlankDeck]},
		{Name: "Lifetime Warranty", Amount: m.RoofSqFt * r[models.KeyWarranty]},
		{Name: "Underlayment", Amount: m.RoofSqFt * r[models.KeyUnderDeckDefense]},
		{Name: "Ridge Vent", Amount: m.RidgeLF * r[models.KeyRidgeVent]},
		{Name: "Intake Vent", Amount: m.IntakeLF * r[models.KeyRidgeVent]},
		{Name: "Starter Strip", Amount: m.Perimeter() * r[models.KeyStarter]},
		{Name: "Drip Edge", Amount: m.Perimeter() * r[models.KeyDripEdge]},
		{Name: "Nails", Amount: m.RoofSqFt * r[models.KeyNails]},
		{Name: "Solar Fan", Amount: float64(m.SolarFans) * r[models.KeySolarFan]},
		{Name: "Power Fan", Amount: float64(m.PowerFans) * r[models.KeyPowerFan]},
		{Name: "Broan Vent", Amount: float64(m.BroanVents) * r[models.KeyBroanVent]},
		{Name: "Pipe Boot", Amount: float64(m.PipeBoots) * r[models.KeyPipeBoot]},
		{Name: "Chimney Flashing", Amount: float64(m.Chimney) * r[models.KeyFlashing] * 2.5},
		{Name: "Flashing", Amount: r[models.KeyFlashing]},
		{Name: "Skylights", Amount: skylightTotal},
	}

	labor := m.RoofSqFt*r[models.KeyLabor] +
		m.AddtlTearoffSF*r[models.KeyAddtlTearoff] +
		m.SteepSlopeSF*r[models.KeySteepSlope]

	result := &models.EstimateResult{
		CustomerName:         m.CustomerName,
		CustomerAddress:      m.CustomerAddress,
		CustomerPhone:        m.CustomerPhone,
		CustomerEmail:        m.CustomerEmail,
		RoofSqFt:             m.RoofSqFt,
		DripEdgeColor:        m.DripEdgeColor,
		StepFlashingColor:    m.StepFlashingColor,
		ApronFlashingColor:   m.ApronFlashingColor,
		ChimneyFlashingColor: m.ChimneyFlashingColor,
		CreatedAt:            time.Now(),
	}

	selected := map[string]bool{}
	for _, k := range m.SelectedTiers {
		selected[k] = true
	}

	for _, def := range models.Tiers {
		// Tier-specific lines come first (Shingles, then shared, then the
		// two tier-specific accessory lines), matching the Excel ordering.
		lines := []models.LineItem{
			{Name: "Shingles", Amount: m.RoofSqFt * r[def.ShingleKey]},
		}
		lines = append(lines, shared...)
		lines = append(lines,
			models.LineItem{Name: "Ice & Water Shield", Amount: ((m.EavesLF * 2) + m.ValleysLF + m.IntakeLF) * r[def.IceWaterKey]},
			models.LineItem{Name: "Hip & Ridge", Amount: (m.RidgeLF + m.HipsLF) * r[def.HipRidgeKey]},
		)

		var matCost float64
		for _, li := range lines {
			matCost += li.Amount
		}

		subtotal := matCost + labor
		total := subtotal / margin
		tr := models.TierResult{
			Key:          def.Key,
			Name:         def.Name,
			ShingleName:  c.rateLabel(def.ShingleKey),
			Lines:        lines,
			MaterialCost: round2(matCost),
			LaborCost:    round2(labor),
			Subtotal:     round2(subtotal),
			Overhead:     round2(total - subtotal),
			TotalPrice:   round2(total),
			Selected:     selected[def.Key],
		}
		if m.RoofSqFt > 0 {
			tr.PricePerSqFt = round2(total / m.RoofSqFt)
		}
		result.Tiers = append(result.Tiers, tr)
	}

	result.OrderList = c.orderList(m)
	return result, nil
}

// orderList converts measurements into physical supplier quantities.
//
//	Shingles     = (roof/100) * 3            bundles
//	Underlayment = (roof/100) / 10           rolls
//	Ice & Water  = (eaves*2 + valleys + intake) / 66  rolls
//	Ridge Vent   = ridge / 4                 pieces
//	Hip & Ridge  = (hips + ridge) / 33       bundles
//	Intake Vent  = intake / 4                pieces
//	Starter      = (eaves + rakes) / 105     bundles
//	Drip Edge    = (eaves + rakes) / 10      sticks
//	Nails        = (roof/100) / 15           boxes
func (c *RoofingCalculator) orderList(m models.RoofMeasurements) []models.OrderItem {
	iw := (m.EavesLF * 2) + m.ValleysLF + m.IntakeLF
	return []models.OrderItem{
		{Name: "Shingles", Qty: ceil1(m.RoofSqFt / 100 * 3), Unit: "bundles"},
		{Name: "Underlayment", Qty: ceil1(m.RoofSqFt / 100 / 10), Unit: "rolls"},
		{Name: "Ice & Water Shield", Qty: ceil1(iw / 66), Unit: "rolls"},
		{Name: "Ridge Vent", Qty: ceil1(m.RidgeLF / 4), Unit: "pieces"},
		{Name: "Hip & Ridge Shingle", Qty: ceil1((m.HipsLF + m.RidgeLF) / 33), Unit: "bundles"},
		{Name: "Intake Vent", Qty: ceil1(m.IntakeLF / 4), Unit: "pieces"},
		{Name: "Starter Shingle", Qty: ceil1(m.Perimeter() / 105), Unit: "bundles"},
		{Name: "Drip Edge", Qty: ceil1(m.Perimeter() / 10), Unit: "sticks"},
		{Name: "Nails", Qty: ceil1(m.RoofSqFt / 100 / 15), Unit: "boxes"},
	}
}

func (c *RoofingCalculator) rateLabel(key string) string {
	if lbl, ok := c.store.RateLabel(key); ok {
		return lbl
	}
	return key
}

func round2(f float64) float64 { return math.Round(f*100) / 100 }

// ceil1 rounds up to whole units for ordering, but keeps a single decimal so
// the estimator can see fractional coverage before rounding their PO.
func ceil1(f float64) float64 { return math.Round(f*10) / 10 }
