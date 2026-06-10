package handlers

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/datsun80zx/exterior_estimator_0.0.1/internal/calc"
	"github.com/datsun80zx/exterior_estimator_0.0.1/internal/db"
	"github.com/datsun80zx/exterior_estimator_0.0.1/internal/models"
)

type Handler struct {
	store     *db.Store
	templates map[string]*template.Template
	roofCalc  *calc.RoofingCalculator
}

func New(store *db.Store, templateDir string) (*Handler, error) {
	funcMap := template.FuncMap{
		"currency": formatCurrency,
		"qty":      func(f float64) string { return strings.TrimRight(strings.TrimRight(fmt.Sprintf("%.1f", f), "0"), ".") },
		"colors": func() []string {
			return []string{"White", "Brown", "Black", "Bronze", "Almond", "Clay", "Gray", "Copper"}
		},
		"skylightMatrix": func() []models.SkylightOption { return models.SkylightMatrix },
	}

	templates := make(map[string]*template.Template)
	layoutFile := filepath.Join(templateDir, "layout.html")
	partialFiles, _ := filepath.Glob(filepath.Join(templateDir, "partials", "*.html"))

	pageFiles, err := filepath.Glob(filepath.Join(templateDir, "*.html"))
	if err != nil {
		return nil, fmt.Errorf("finding templates: %w", err)
	}

	for _, page := range pageFiles {
		name := filepath.Base(page)
		if name == "layout.html" {
			continue
		}
		files := []string{layoutFile}
		files = append(files, partialFiles...)
		files = append(files, page)
		t, err := template.New(name).Funcs(funcMap).ParseFiles(files...)
		if err != nil {
			return nil, fmt.Errorf("parsing template %s: %w", name, err)
		}
		templates[name] = t
	}
	for _, pf := range partialFiles {
		name := filepath.Base(pf)
		t, err := template.New(name).Funcs(funcMap).ParseFiles(pf)
		if err != nil {
			return nil, fmt.Errorf("parsing partial %s: %w", name, err)
		}
		templates[name] = t
	}

	return &Handler{
		store:     store,
		templates: templates,
		roofCalc:  calc.NewRoofingCalculator(store),
	}, nil
}

func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /", h.handleHome)
	mux.HandleFunc("GET /roofing", h.handleRoofingForm)
	mux.HandleFunc("POST /roofing/calculate", h.handleRoofingCalculate)
	mux.HandleFunc("GET /roofing/proposal/{id}", h.handleProposal)

	// Price book admin
	mux.HandleFunc("GET /admin/pricebook", h.handlePriceBook)
	mux.HandleFunc("POST /admin/pricebook/{id}", h.handleRateUpdate)
}

// --- Pages ---

func (h *Handler) handleHome(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	h.render(w, "home.html", nil)
}

func (h *Handler) handleRoofingForm(w http.ResponseWriter, r *http.Request) {
	h.render(w, "roofing_form.html", map[string]any{
		"Tiers": models.Tiers,
	})
}

func (h *Handler) handleRoofingCalculate(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), 400)
		return
	}

	m := models.RoofMeasurements{
		CustomerName:    r.FormValue("customer_name"),
		CustomerAddress: r.FormValue("customer_address"),
		CustomerPhone:   r.FormValue("customer_phone"),
		CustomerEmail:   r.FormValue("customer_email"),

		RoofSqFt:  parseFloat(r.FormValue("roof_sqft")),
		EavesLF:   parseFloat(r.FormValue("eaves_lf")),
		RakesLF:   parseFloat(r.FormValue("rakes_lf")),
		ValleysLF: parseFloat(r.FormValue("valleys_lf")),
		HipsLF:    parseFloat(r.FormValue("hips_lf")),
		RidgeLF:   parseFloat(r.FormValue("ridge_lf")),
		IntakeLF:  parseFloat(r.FormValue("intake_lf")),

		Chimney:    parseInt(r.FormValue("chimney")),
		PipeBoots:  parseInt(r.FormValue("pipe_boots")),
		Decking:    parseInt(r.FormValue("decking")),
		SolarFans:  parseInt(r.FormValue("solar_fans")),
		PowerFans:  parseInt(r.FormValue("power_fans")),
		BroanVents: parseInt(r.FormValue("broan_vents")),
		Dumpsters:  parseInt(r.FormValue("dumpsters")),
		Skylights:  parseInt(r.FormValue("skylights")),

		PlankDeckLF:    parseFloat(r.FormValue("plank_deck_lf")),
		AddtlTearoffSF: parseFloat(r.FormValue("addtl_tearoff_sf")),
		SteepSlopeSF:   parseFloat(r.FormValue("steep_slope_sf")),

		DripEdgeColor:        r.FormValue("drip_edge_color"),
		StepFlashingColor:    r.FormValue("step_flashing_color"),
		ApronFlashingColor:   r.FormValue("apron_flashing_color"),
		ChimneyFlashingColor: r.FormValue("chimney_flashing_color"),

		SkylightMount: r.FormValue("skylight_mount"),
		SkylightType:  r.FormValue("skylight_type"),
		SkylightSize:  r.FormValue("skylight_size"),
	}

	result, err := h.roofCalc.Calculate(m)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	estimateJSON, _ := json.Marshal(result)
	id, err := h.store.SaveEstimate(result.CustomerName, result.CustomerAddress, "Roofing", string(estimateJSON))
	if err != nil {
		log.Printf("Error saving estimate: %v", err)
	}

	data := map[string]any{"Result": result, "EstimateID": id}
	if r.Header.Get("HX-Request") == "true" {
		h.renderPartial(w, "estimate_result.html", data)
		return
	}
	h.render(w, "estimate_result_page.html", data)
}

func (h *Handler) handleProposal(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	estimateJSON, err := h.store.GetEstimateJSON(id)
	if err != nil {
		http.Error(w, "Estimate not found", 404)
		return
	}

	var result models.EstimateResult
	if err := json.Unmarshal([]byte(estimateJSON), &result); err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	// Tier selection comes from ?tiers=class3,luxury
	if q := r.URL.Query().Get("tiers"); q != "" {
		want := map[string]bool{}
		for _, k := range strings.Split(q, ",") {
			want[strings.TrimSpace(k)] = true
		}
		for i := range result.Tiers {
			result.Tiers[i].Selected = want[result.Tiers[i].Key]
		}
	}

	h.renderPartial(w, "proposal.html", map[string]any{"Result": result})
}

// --- Price book admin ---

func (h *Handler) handlePriceBook(w http.ResponseWriter, r *http.Request) {
	groups, err := h.store.GetRatesByGroup()
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	h.render(w, "admin_pricebook.html", map[string]any{"Groups": groups})
}

func (h *Handler) handleRateUpdate(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	id := parseInt64(r.PathValue("id"))
	amount := parseFloat(r.FormValue("amount"))
	if err := h.store.UpdateRateAmount(id, amount); err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	// Return the updated amount cell for HTMX swap.
	rate, err := h.store.GetRate(id)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprintf(w, `<span class="saved">Saved · %s</span>`, formatCurrency(rate.Amount))
}

// --- Render helpers ---

func (h *Handler) render(w http.ResponseWriter, name string, data any) {
	t, ok := h.templates[name]
	if !ok {
		log.Printf("Template not found: %s", name)
		http.Error(w, "Internal Server Error", 500)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := t.ExecuteTemplate(w, "layout.html", data); err != nil {
		log.Printf("Template error (%s): %v", name, err)
		http.Error(w, "Internal Server Error", 500)
	}
}

func (h *Handler) renderPartial(w http.ResponseWriter, name string, data any) {
	t, ok := h.templates[name]
	if !ok {
		log.Printf("Partial template not found: %s", name)
		http.Error(w, "Internal Server Error", 500)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := t.ExecuteTemplate(w, name, data); err != nil {
		log.Printf("Partial template error (%s): %v", name, err)
		http.Error(w, "Internal Server Error", 500)
	}
}

// formatCurrency renders a dollar amount with thousands separators, e.g.
// 31172.44 -> "$31,172.44". Handles negatives.
func formatCurrency(f float64) string {
	neg := f < 0
	if neg {
		f = -f
	}
	s := fmt.Sprintf("%.2f", f) // "31172.44"
	intPart, frac := s, ""
	if dot := strings.IndexByte(s, '.'); dot >= 0 {
		intPart, frac = s[:dot], s[dot:]
	}
	var b strings.Builder
	n := len(intPart)
	for i, c := range intPart {
		if i > 0 && (n-i)%3 == 0 {
			b.WriteByte(',')
		}
		b.WriteRune(c)
	}
	sign := ""
	if neg {
		sign = "-"
	}
	return sign + "$" + b.String() + frac
}

func parseFloat(s string) float64 {
	f, _ := strconv.ParseFloat(strings.TrimSpace(s), 64)
	return f
}

func parseInt(s string) int {
	i, _ := strconv.Atoi(strings.TrimSpace(s))
	return i
}

func parseInt64(s string) int64 {
	i, _ := strconv.ParseInt(strings.TrimSpace(s), 10, 64)
	return i
}
