package calc

import (
	"fmt"
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

func (c *RoofingCalculator) Calculate(m models.RoofMeasurements) (*models.EstimateResult, error) {
	// Resolve pitch multiplier
	pitchMult, ok := models.PitchMultipliers[m.Pitch]
	if !ok {
		pitchMult = 1.0
	}
	m.PitchMultiplier = pitchMult

	// Waste factor as decimal
	wasteFactor := 1.0 + (m.WasteFactorPct / 100.0)

	// Adjusted area with waste
	adjustedArea := m.TotalAreaSqFt * wasteFactor
	totalSquares := adjustedArea / 100.0

	result := &models.EstimateResult{
		CustomerName:    m.CustomerName,
		CustomerAddress: m.CustomerAddress,
		CustomerPhone:   m.CustomerPhone,
		CustomerEmail:   m.CustomerEmail,
		ProjectType:     "Roofing",
		TotalSquares:    math.Round(totalSquares*10) / 10,
		Pitch:           m.Pitch,
		WasteFactorPct:  m.WasteFactorPct,
		CreatedAt:       time.Now(),
	}

	// --- SHINGLES ---
	if m.ShingleMaterialID > 0 {
		mat, err := c.store.GetMaterial(m.ShingleMaterialID)
		if err != nil {
			return nil, fmt.Errorf("getting shingle material: %w", err)
		}
		qty := int(math.Ceil(adjustedArea / mat.CoveragePerUnit))
		result.Materials = append(result.Materials, models.MaterialLineItem{
			MaterialName: mat.Name,
			Brand:        mat.Brand,
			ProductLine:  mat.ProductLine,
			Quantity:     qty,
			Unit:         mat.Unit,
			CostEach:     mat.CostPerUnit,
			CostTotal:    float64(qty) * mat.CostPerUnit,
			PriceEach:    mat.PricePerUnit,
			PriceTotal:   float64(qty) * mat.PricePerUnit,
			Notes:        fmt.Sprintf("%.1f squares + %.0f%% waste", m.TotalAreaSqFt/100, m.WasteFactorPct),
		})
	}

	// --- UNDERLAYMENT ---
	if m.UnderlaymentMaterialID > 0 {
		mat, err := c.store.GetMaterial(m.UnderlaymentMaterialID)
		if err != nil {
			return nil, fmt.Errorf("getting underlayment material: %w", err)
		}
		// Underlayment covers the full roof area
		qty := int(math.Ceil(m.TotalAreaSqFt / mat.CoveragePerUnit))
		result.Materials = append(result.Materials, models.MaterialLineItem{
			MaterialName: mat.Name,
			Brand:        mat.Brand,
			ProductLine:  mat.ProductLine,
			Quantity:     qty,
			Unit:         mat.Unit,
			CostEach:     mat.CostPerUnit,
			CostTotal:    float64(qty) * mat.CostPerUnit,
			PriceEach:    mat.PricePerUnit,
			PriceTotal:   float64(qty) * mat.PricePerUnit,
			Notes:        fmt.Sprintf("%.0f sq ft coverage needed", m.TotalAreaSqFt),
		})
	}

	// --- ICE & WATER SHIELD ---
	if m.IceWaterMaterialID > 0 {
		mat, err := c.store.GetMaterial(m.IceWaterMaterialID)
		if err != nil {
			return nil, fmt.Errorf("getting ice & water material: %w", err)
		}
		// Ice & water runs along eaves (3ft x 2 = 6ft up from eave) + valleys
		iceWaterArea := (m.EaveLengthFt * 6) + (m.ValleyLengthFt * 3)
		qty := int(math.Ceil(iceWaterArea / mat.CoveragePerUnit))
		if qty < 1 {
			qty = 1
		}
		result.Materials = append(result.Materials, models.MaterialLineItem{
			MaterialName: mat.Name,
			Brand:        mat.Brand,
			ProductLine:  mat.ProductLine,
			Quantity:     qty,
			Unit:         mat.Unit,
			CostEach:     mat.CostPerUnit,
			CostTotal:    float64(qty) * mat.CostPerUnit,
			PriceEach:    mat.PricePerUnit,
			PriceTotal:   float64(qty) * mat.PricePerUnit,
			Notes:        fmt.Sprintf("Eaves: %.0f' × 6' + Valleys: %.0f' × 3'", m.EaveLengthFt, m.ValleyLengthFt),
		})
	}

	// --- DRIP EDGE ---
	if m.DripEdgeMaterialID > 0 {
		mat, err := c.store.GetMaterial(m.DripEdgeMaterialID)
		if err != nil {
			return nil, fmt.Errorf("getting drip edge material: %w", err)
		}
		// Drip edge along eaves + rakes
		totalDripLF := m.EaveLengthFt + m.RakeLengthFt
		qty := int(math.Ceil(totalDripLF / mat.CoveragePerUnit))
		result.Materials = append(result.Materials, models.MaterialLineItem{
			MaterialName: mat.Name,
			Brand:        mat.Brand,
			ProductLine:  mat.ProductLine,
			Quantity:     qty,
			Unit:         mat.Unit,
			CostEach:     mat.CostPerUnit,
			CostTotal:    float64(qty) * mat.CostPerUnit,
			PriceEach:    mat.PricePerUnit,
			PriceTotal:   float64(qty) * mat.PricePerUnit,
			Notes:        fmt.Sprintf("Eave: %.0f' + Rake: %.0f' = %.0f LF", m.EaveLengthFt, m.RakeLengthFt, totalDripLF),
		})
	}

	// --- RIDGE CAP ---
	if m.RidgeCapMaterialID > 0 {
		mat, err := c.store.GetMaterial(m.RidgeCapMaterialID)
		if err != nil {
			return nil, fmt.Errorf("getting ridge cap material: %w", err)
		}
		totalRidgeLF := m.RidgeLengthFt + m.HipLengthFt
		qty := int(math.Ceil(totalRidgeLF / mat.CoveragePerUnit))
		if qty < 1 && totalRidgeLF > 0 {
			qty = 1
		}
		result.Materials = append(result.Materials, models.MaterialLineItem{
			MaterialName: mat.Name,
			Brand:        mat.Brand,
			ProductLine:  mat.ProductLine,
			Quantity:     qty,
			Unit:         mat.Unit,
			CostEach:     mat.CostPerUnit,
			CostTotal:    float64(qty) * mat.CostPerUnit,
			PriceEach:    mat.PricePerUnit,
			PriceTotal:   float64(qty) * mat.PricePerUnit,
			Notes:        fmt.Sprintf("Ridge: %.0f' + Hips: %.0f' = %.0f LF", m.RidgeLengthFt, m.HipLengthFt, totalRidgeLF),
		})
	}

	// --- STARTER STRIP ---
	if m.StarterMaterialID > 0 {
		mat, err := c.store.GetMaterial(m.StarterMaterialID)
		if err != nil {
			return nil, fmt.Errorf("getting starter material: %w", err)
		}
		// Starter runs along eaves + rakes
		totalStarterLF := m.EaveLengthFt + m.RakeLengthFt
		qty := int(math.Ceil(totalStarterLF / mat.CoveragePerUnit))
		result.Materials = append(result.Materials, models.MaterialLineItem{
			MaterialName: mat.Name,
			Brand:        mat.Brand,
			ProductLine:  mat.ProductLine,
			Quantity:     qty,
			Unit:         mat.Unit,
			CostEach:     mat.CostPerUnit,
			CostTotal:    float64(qty) * mat.CostPerUnit,
			PriceEach:    mat.PricePerUnit,
			PriceTotal:   float64(qty) * mat.PricePerUnit,
			Notes:        fmt.Sprintf("%.0f LF total (eave + rake)", totalStarterLF),
		})
	}

	// --- PIPE BOOTS ---
	if m.NumPipeBoots > 0 {
		// Grab a default pipe boot
		mat, _ := c.findMaterialByName("roofing", "Pipe Boot")
		if mat.ID > 0 {
			result.Materials = append(result.Materials, models.MaterialLineItem{
				MaterialName: mat.Name,
				Brand:        mat.Brand,
				Quantity:     m.NumPipeBoots,
				Unit:         "piece",
				CostEach:     mat.CostPerUnit,
				CostTotal:    float64(m.NumPipeBoots) * mat.CostPerUnit,
				PriceEach:    mat.PricePerUnit,
				PriceTotal:   float64(m.NumPipeBoots) * mat.PricePerUnit,
			})
		}
	}

	// --- EXHAUST VENTS ---
	if m.NumExhaustVents > 0 {
		mat, _ := c.findMaterialByName("roofing", "Exhaust Vent")
		if mat.ID > 0 {
			result.Materials = append(result.Materials, models.MaterialLineItem{
				MaterialName: mat.Name,
				Brand:        mat.Brand,
				Quantity:     m.NumExhaustVents,
				Unit:         "piece",
				CostEach:     mat.CostPerUnit,
				CostTotal:    float64(m.NumExhaustVents) * mat.CostPerUnit,
				PriceEach:    mat.PricePerUnit,
				PriceTotal:   float64(m.NumExhaustVents) * mat.PricePerUnit,
			})
		}
	}

	// --- NAILS ---
	// ~320 nails per square (4 nails per shingle, 80 per bundle, 3 bundles per sq)
	// A box of 7200 covers ~22 squares
	nailSquares := totalSquares
	nailBoxes := int(math.Ceil(nailSquares / 22.0))
	mat, _ := c.findMaterialByName("roofing", "Coil Roofing Nails")
	if mat.ID > 0 {
		result.Materials = append(result.Materials, models.MaterialLineItem{
			MaterialName: mat.Name,
			Brand:        mat.Brand,
			Quantity:     nailBoxes,
			Unit:         "box",
			CostEach:     mat.CostPerUnit,
			CostTotal:    float64(nailBoxes) * mat.CostPerUnit,
			PriceEach:    mat.PricePerUnit,
			PriceTotal:   float64(nailBoxes) * mat.PricePerUnit,
			Notes:        "~320 nails/sq, box of 7200 = ~22 squares",
		})
	}

	// --- Calculate Totals ---
	for _, item := range result.Materials {
		result.TotalMaterialCost += item.CostTotal
		result.TotalMaterialPrice += item.PriceTotal
	}

	// Labor
	lr, _ := c.store.GetLaborRate(models.CategoryRoofing)
	result.LaborCost = math.Round(totalSquares*lr.RatePerSq*100) / 100

	// Tear-off
	if m.LayersToRemove > 0 {
		result.TearOffCost = math.Round(totalSquares*50*float64(m.LayersToRemove)*100) / 100
	}

	result.TotalProjectCost = result.TotalMaterialCost + result.LaborCost + result.TearOffCost
	result.TotalProjectPrice = result.TotalMaterialPrice + result.LaborCost + result.TearOffCost
	result.Profit = result.TotalProjectPrice - result.TotalProjectCost
	if result.TotalProjectPrice > 0 {
		result.MarginPct = math.Round((result.Profit/result.TotalProjectPrice)*10000) / 100
	}

	return result, nil
}

func (c *RoofingCalculator) findMaterialByName(category, nameContains string) (models.Material, error) {
	var m models.Material
	err := c.store.DB.QueryRow(`
		SELECT id, category, brand, product_line, name, unit,
		       coverage_per_unit, cost_per_unit, price_per_unit, is_active
		FROM materials
		WHERE category = ? AND name LIKE ? AND is_active = 1
		LIMIT 1
	`, category, "%"+nameContains+"%").Scan(&m.ID, &m.Category, &m.Brand, &m.ProductLine,
		&m.Name, &m.Unit, &m.CoveragePerUnit, &m.CostPerUnit, &m.PricePerUnit, &m.IsActive)
	return m, err
}
