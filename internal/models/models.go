package models

import "time"

// =====================================================================
// PRICE BOOK
// =====================================================================
// The Excel price book is a flat list of editable rates. Each rate is a
// dollar amount applied per sqft, per linear foot, or per each, plus a
// couple of "setting" rows (business margin). Everything in the calc
// engine references these by their stable Key.

type RateGroup string

const (
	GroupShingle  RateGroup = "shingle"
	GroupHipRidge RateGroup = "hip_ridge"
	GroupStarter  RateGroup = "starter"
	GroupIceWater RateGroup = "ice_water"
	GroupUnderlay RateGroup = "underlayment"
	GroupMisc     RateGroup = "misc"
	GroupLabor    RateGroup = "labor"
	GroupSetting  RateGroup = "setting"
)

// Rate is one editable line in the price book.
type Rate struct {
	ID     int64     `json:"id"`
	Key    string    `json:"key"`    // stable identifier referenced by the calc engine
	Label  string    `json:"label"`  // human-readable name
	Group  RateGroup `json:"group"`  // bucket for the admin UI
	Unit   string    `json:"unit"`   // "/sqft", "/lf", "/ea", "/sheet", "factor"
	Amount float64   `json:"amount"` // dollar amount (or factor for settings)
}

// Stable rate keys. These mirror column B of the Excel price book.
const (
	// Shingles ($/sqft)
	KeyShingleOakridge   = "shingle_oakridge"
	KeyShingleDuration   = "shingle_duration"
	KeyShingleDesigner   = "shingle_designer"
	KeyShingleFlex       = "shingle_flex"
	KeyShingleBrava      = "shingle_brava"
	KeyShingleGrandManor = "shingle_grand_manor"

	// Hip & Ridge ($/lf)
	KeyHRProEdge     = "hr_proedge"
	KeyHRProEdgeFlex = "hr_proedge_flex"
	KeyHRBrava       = "hr_brava"
	KeyHRGrandManor  = "hr_grand_manor"

	// Starter ($/lf)
	KeyStarter = "starter"

	// Ice & Water ($/lf)
	KeyIWWeatherlockG    = "iw_weatherlock_g"
	KeyIWWeatherlockMat  = "iw_weatherlock_mat"
	KeyIWWeatherlockFlex = "iw_weatherlock_flex"

	// Underlayment ($/sqft)
	KeyUnderProArmor    = "under_proarmor"
	KeyUnderDeckDefense = "under_deckdefense"
	KeyUnderRhinoRoof   = "under_rhinoroof"

	// Misc
	KeyPipeBoot  = "pipeboot"   // $/ea
	KeyNails     = "nails"      // $/sqft
	KeyFlashing  = "flashing"   // $/ea (flat line + chimney x2.5)
	KeyDripEdge  = "drip_edge"  // $/lf
	KeyRidgeVent = "ridge_vent" // $/lf (used for ridge vent AND intake vent)
	KeyOSB       = "osb"        // $/sheet
	KeyWarranty  = "warranty"   // $/sqft
	KeyPlankDeck = "plank_deck" // $/lf
	KeySolarFan  = "solar_fan"  // $/ea
	KeyPowerFan  = "power_fan"  // $/ea
	KeyBroanVent = "broan_vent" // $/ea
	KeyDumpster  = "dumpster"   // $/ea (per 15 sq)
	KeySkylight  = "skylight"   // $/ea (flat default)

	// Labor
	KeyLabor          = "labor"           // $/sqft
	KeyLaborSpecialty = "labor_specialty" // $/sqft
	KeyAddtlTearoff   = "addtl_tearoff"   // $/sqft
	KeySteepSlope     = "steep_slope"     // $/sqft

	// Settings
	KeyBusinessMargin = "business_margin" // divisor, e.g. 0.85
)

// =====================================================================
// TIERS
// =====================================================================
// The four packages differ only by shingle, hip & ridge, and ice & water.
// Everything else (underlayment, starter, accessories, add-ons) is shared.

type TierDef struct {
	Key         string
	Name        string
	ShingleKey  string
	HipRidgeKey string
	IceWaterKey string
}

// Tiers reproduces the Excel package columns (Class 3 / Class 4 / Luxury /
// Faux Slate) exactly.
var Tiers = []TierDef{
	{Key: "class3", Name: "Class 3", ShingleKey: KeyShingleDuration, HipRidgeKey: KeyHRProEdge, IceWaterKey: KeyIWWeatherlockG},
	{Key: "class4", Name: "Class 4", ShingleKey: KeyShingleFlex, HipRidgeKey: KeyHRProEdgeFlex, IceWaterKey: KeyIWWeatherlockFlex},
	{Key: "luxury", Name: "Luxury", ShingleKey: KeyShingleGrandManor, HipRidgeKey: KeyHRGrandManor, IceWaterKey: KeyIWWeatherlockFlex},
	{Key: "faux", Name: "Faux Slate", ShingleKey: KeyShingleBrava, HipRidgeKey: KeyHRBrava, IceWaterKey: KeyIWWeatherlockFlex},
}

// =====================================================================
// MEASUREMENTS (the variables the estimator enters per job)
// =====================================================================

type RoofMeasurements struct {
	CustomerName    string `json:"customer_name"`
	CustomerAddress string `json:"customer_address"`
	CustomerPhone   string `json:"customer_phone"`
	CustomerEmail   string `json:"customer_email"`

	// Areas / lengths
	RoofSqFt  float64 `json:"roof_sqft"` // measured roof surface area (already pitched)
	EavesLF   float64 `json:"eaves_lf"`
	RakesLF   float64 `json:"rakes_lf"`
	ValleysLF float64 `json:"valleys_lf"`
	HipsLF    float64 `json:"hips_lf"`
	RidgeLF   float64 `json:"ridge_lf"`
	IntakeLF  float64 `json:"intake_lf"`

	// Counts
	Chimney    int `json:"chimney"`
	PipeBoots  int `json:"pipe_boots"`
	Decking    int `json:"decking"` // sheets of OSB
	SolarFans  int `json:"solar_fans"`
	PowerFans  int `json:"power_fans"`
	BroanVents int `json:"broan_vents"`
	Dumpsters  int `json:"dumpsters"`
	Skylights  int `json:"skylights"`

	PlankDeckLF    float64 `json:"plank_deck_lf"`
	AddtlTearoffSF float64 `json:"addtl_tearoff_sf"`
	SteepSlopeSF   float64 `json:"steep_slope_sf"`

	// Colors
	DripEdgeColor        string `json:"drip_edge_color"`
	StepFlashingColor    string `json:"step_flashing_color"`
	ApronFlashingColor   string `json:"apron_flashing_color"`
	ChimneyFlashingColor string `json:"chimney_flashing_color"`

	// Optional detailed skylight selection (overrides flat skylight price)
	SkylightMount string `json:"skylight_mount"` // "curb" | "deck" | "" (flat)
	SkylightType  string `json:"skylight_type"`  // "solar" | "electric" | "manual" | "fixed"
	SkylightSize  string `json:"skylight_size"`  // size key from the matrix

	// Which tiers to surface on the customer proposal
	SelectedTiers []string `json:"selected_tiers"`
}

// Perimeter is eaves + rakes (used for starter / drip edge order quantities).
func (m RoofMeasurements) Perimeter() float64 { return m.EavesLF + m.RakesLF }

// =====================================================================
// OUTPUT
// =====================================================================

type LineItem struct {
	Name   string  `json:"name"`
	Amount float64 `json:"amount"`
}

type TierResult struct {
	Key          string     `json:"key"`
	Name         string     `json:"name"`
	ShingleName  string     `json:"shingle_name"`
	Lines        []LineItem `json:"lines"`
	MaterialCost float64    `json:"material_cost"`
	LaborCost    float64    `json:"labor_cost"`
	Subtotal     float64    `json:"subtotal"`
	Overhead     float64    `json:"overhead"`
	TotalPrice   float64    `json:"total_price"`
	PricePerSqFt float64    `json:"price_per_sqft"`
	Selected     bool       `json:"selected"` // shown on customer proposal
}

// OrderItem is a physical-quantity line for the supplier order (shared
// across tiers; depends only on measurements).
type OrderItem struct {
	Name string  `json:"name"`
	Qty  float64 `json:"qty"`
	Unit string  `json:"unit"`
}

type EstimateResult struct {
	CustomerName    string `json:"customer_name"`
	CustomerAddress string `json:"customer_address"`
	CustomerPhone   string `json:"customer_phone"`
	CustomerEmail   string `json:"customer_email"`

	RoofSqFt float64 `json:"roof_sqft"`

	DripEdgeColor        string `json:"drip_edge_color"`
	StepFlashingColor    string `json:"step_flashing_color"`
	ApronFlashingColor   string `json:"apron_flashing_color"`
	ChimneyFlashingColor string `json:"chimney_flashing_color"`

	Tiers     []TierResult `json:"tiers"`
	OrderList []OrderItem  `json:"order_list"`

	CreatedAt time.Time `json:"created_at"`
}

// SelectedTiers returns only the tiers flagged for the customer proposal.
func (e EstimateResult) SelectedTierResults() []TierResult {
	var out []TierResult
	for _, t := range e.Tiers {
		if t.Selected {
			out = append(out, t)
		}
	}
	if len(out) == 0 {
		return e.Tiers // fall back to showing all
	}
	return out
}

// =====================================================================
// SKYLIGHT MATRIX (reference pricing by mount x type x size)
// =====================================================================

type SkylightOption struct {
	Mount string // "curb" | "deck"
	Type  string // "solar" | "electric" | "manual" | "fixed"
	Size  string // e.g. "34x50"
	Label string // display label
	Price float64
}

// SkylightMatrix mirrors the Excel skylight table (U2:AF8).
var SkylightMatrix = []SkylightOption{
	{"curb", "solar", "34x50", "Curb · Solar · up to 34x50", 2206},
	{"curb", "solar", "50x50", "Curb · Solar · up to 50x50", 2328},
	{"curb", "electric", "34x50", "Curb · Electric · up to 34x50", 2206},
	{"curb", "electric", "50x50", "Curb · Electric · up to 50x50", 2328},
	{"curb", "manual", "34x50", "Curb · Manual · up to 34x50", 1494},
	{"curb", "manual", "50x50", "Curb · Manual · up to 50x50", 1642},
	{"curb", "fixed", "34x50", "Curb · Fixed · up to 34x50", 868},
	{"curb", "fixed", "46x72", "Curb · Fixed · up to 46x72", 1327},
	{"deck", "solar", "21x72", "Deck · Solar · up to 21x72", 2050},
	{"deck", "solar", "31x55", "Deck · Solar · up to 31x55", 2300},
	{"deck", "electric", "21x72", "Deck · Electric · up to 21x72", 2050},
	{"deck", "electric", "31x55", "Deck · Electric · up to 31x55", 2300},
	{"deck", "manual", "21x72", "Deck · Manual · up to 21x72", 1200},
	{"deck", "manual", "31x55", "Deck · Manual · up to 31x55", 1300},
	{"deck", "fixed", "21x72", "Deck · Fixed · up to 21x72", 1050},
	{"deck", "fixed", "31x55", "Deck · Fixed · up to 31x55", 1150},
}

// LookupSkylight returns the matrix price for a mount/type/size, ok=false if
// no match (caller should fall back to the flat skylight rate).
func LookupSkylight(mount, typ, size string) (float64, bool) {
	for _, o := range SkylightMatrix {
		if o.Mount == mount && o.Type == typ && o.Size == size {
			return o.Price, true
		}
	}
	return 0, false
}
