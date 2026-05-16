package models

import "time"

// --- Material & Pricing ---

type MaterialCategory string

const (
	CategoryRoofing MaterialCategory = "roofing"
	CategorySiding  MaterialCategory = "siding"
	CategoryGutters MaterialCategory = "gutters"
)

type Material struct {
	ID              int64            `json:"id"`
	Category        MaterialCategory `json:"category"`
	Brand           string           `json:"brand"`
	ProductLine     string           `json:"product_line"`
	Name            string           `json:"name"`
	Unit            string           `json:"unit"`              // "bundle", "roll", "piece", "box", "linear_ft"
	CoveragePerUnit float64          `json:"coverage_per_unit"` // sq ft per unit, or linear ft per unit
	CostPerUnit     float64          `json:"cost_per_unit"`     // your cost
	PricePerUnit    float64          `json:"price_per_unit"`    // sell price (or calculate from margin)
	IsActive        bool             `json:"is_active"`
	CreatedAt       time.Time        `json:"created_at"`
	UpdatedAt       time.Time        `json:"updated_at"`
}

// --- Roofing Measurements ---

type RoofMeasurements struct {
	CustomerName    string `json:"customer_name"`
	CustomerAddress string `json:"customer_address"`
	CustomerPhone   string `json:"customer_phone"`
	CustomerEmail   string `json:"customer_email"`

	TotalAreaSqFt   float64 `json:"total_area_sqft"`  // total roof area
	Pitch           string  `json:"pitch"`            // e.g. "6/12"
	PitchMultiplier float64 `json:"pitch_multiplier"` // calculated from pitch

	RidgeLengthFt  float64 `json:"ridge_length_ft"`
	HipLengthFt    float64 `json:"hip_length_ft"`
	ValleyLengthFt float64 `json:"valley_length_ft"`
	EaveLengthFt   float64 `json:"eave_length_ft"`
	RakeLengthFt   float64 `json:"rake_length_ft"`

	NumPipeBoots    int `json:"num_pipe_boots"`
	NumExhaustVents int `json:"num_exhaust_vents"`
	LayersToRemove  int `json:"layers_to_remove"`

	WasteFactorPct float64 `json:"waste_factor_pct"` // e.g. 10, 15, 20

	// Selected material IDs
	ShingleMaterialID      int64 `json:"shingle_material_id"`
	UnderlaymentMaterialID int64 `json:"underlayment_material_id"`
	IceWaterMaterialID     int64 `json:"ice_water_material_id"`
	DripEdgeMaterialID     int64 `json:"drip_edge_material_id"`
	RidgeCapMaterialID     int64 `json:"ridge_cap_material_id"`
	StarterMaterialID      int64 `json:"starter_material_id"`
}

// --- Calculated Output ---

type MaterialLineItem struct {
	MaterialName string  `json:"material_name"`
	Brand        string  `json:"brand"`
	ProductLine  string  `json:"product_line"`
	Quantity     int     `json:"quantity"`
	Unit         string  `json:"unit"`
	CostEach     float64 `json:"cost_each"`
	CostTotal    float64 `json:"cost_total"`
	PriceEach    float64 `json:"price_each"`
	PriceTotal   float64 `json:"price_total"`
	Notes        string  `json:"notes"`
}

type EstimateResult struct {
	// Customer info
	CustomerName    string `json:"customer_name"`
	CustomerAddress string `json:"customer_address"`
	CustomerPhone   string `json:"customer_phone"`
	CustomerEmail   string `json:"customer_email"`

	// Project summary
	ProjectType    string  `json:"project_type"` // "Roofing", "Siding", "Gutters"
	TotalSquares   float64 `json:"total_squares"`
	Pitch          string  `json:"pitch"`
	WasteFactorPct float64 `json:"waste_factor_pct"`

	// Material breakdown
	Materials []MaterialLineItem `json:"materials"`

	// Totals
	TotalMaterialCost  float64 `json:"total_material_cost"`
	TotalMaterialPrice float64 `json:"total_material_price"`
	LaborCost          float64 `json:"labor_cost"`
	TearOffCost        float64 `json:"tear_off_cost"`
	TotalProjectCost   float64 `json:"total_project_cost"`
	TotalProjectPrice  float64 `json:"total_project_price"`
	Profit             float64 `json:"profit"`
	MarginPct          float64 `json:"margin_pct"`

	CreatedAt time.Time `json:"created_at"`
}

// --- Labor Rates (configurable) ---

type LaborRate struct {
	ID          int64            `json:"id"`
	Category    MaterialCategory `json:"category"`
	Description string           `json:"description"`
	RatePerSq   float64          `json:"rate_per_sq"`   // per square for roofing
	RatePerLF   float64          `json:"rate_per_lf"`   // per linear foot for gutters
	RatePerSqFt float64          `json:"rate_per_sqft"` // per sq ft for siding
	IsActive    bool             `json:"is_active"`
}

// PitchMultipliers maps common pitches to their area multiplier
var PitchMultipliers = map[string]float64{
	"flat":  1.000,
	"1/12":  1.003,
	"2/12":  1.014,
	"3/12":  1.031,
	"4/12":  1.054,
	"5/12":  1.083,
	"6/12":  1.118,
	"7/12":  1.158,
	"8/12":  1.202,
	"9/12":  1.250,
	"10/12": 1.302,
	"11/12": 1.357,
	"12/12": 1.414,
	"13/12": 1.474,
	"14/12": 1.537,
}
