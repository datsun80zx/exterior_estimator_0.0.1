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
		"currency": func(f float64) string {
			return fmt.Sprintf("$%.2f", f)
		},
		"multiply": func(a, b float64) float64 {
			return a * b
		},
		"pitches": func() []string {
			return []string{
				"flat", "1/12", "2/12", "3/12", "4/12", "5/12", "6/12",
				"7/12", "8/12", "9/12", "10/12", "11/12", "12/12", "13/12", "14/12",
			}
		},
		"seq": func(n int) []int {
			s := make([]int, n)
			for i := range s {
				s[i] = i
			}
			return s
		},
	}

	templates := make(map[string]*template.Template)

	// Parse layout as the base template
	layoutFile := filepath.Join(templateDir, "layout.html")

	// Parse partials (shared fragments like estimate_result.html)
	partialFiles, _ := filepath.Glob(filepath.Join(templateDir, "partials", "*.html"))

	// Each page template gets: layout + all partials + itself
	pageFiles, err := filepath.Glob(filepath.Join(templateDir, "*.html"))
	if err != nil {
		return nil, fmt.Errorf("finding templates: %w", err)
	}

	for _, page := range pageFiles {
		name := filepath.Base(page)
		if name == "layout.html" {
			continue
		}

		// Build the file list: layout first, then partials, then the page
		files := []string{layoutFile}
		files = append(files, partialFiles...)
		files = append(files, page)

		t, err := template.New(name).Funcs(funcMap).ParseFiles(files...)
		if err != nil {
			return nil, fmt.Errorf("parsing template %s: %w", name, err)
		}
		templates[name] = t
	}

	// Also parse partials standalone (for HTMX responses)
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
	// Pages
	mux.HandleFunc("GET /", h.handleHome)
	mux.HandleFunc("GET /roofing", h.handleRoofingForm)
	mux.HandleFunc("POST /roofing/calculate", h.handleRoofingCalculate)
	mux.HandleFunc("GET /roofing/proposal/{id}", h.handleProposal)

	// Material admin
	mux.HandleFunc("GET /admin/materials", h.handleMaterialsList)
	mux.HandleFunc("GET /admin/materials/new", h.handleMaterialForm)
	mux.HandleFunc("GET /admin/materials/{id}/edit", h.handleMaterialEdit)
	mux.HandleFunc("POST /admin/materials", h.handleMaterialSave)
	mux.HandleFunc("DELETE /admin/materials/{id}", h.handleMaterialDelete)

	// HTMX partials
	mux.HandleFunc("GET /htmx/shingles-by-brand", h.handleShinglesByBrand)
}

// --- Page Handlers ---

func (h *Handler) handleHome(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	h.render(w, "home.html", nil)
}

func (h *Handler) handleRoofingForm(w http.ResponseWriter, r *http.Request) {
	materials, err := h.store.GetMaterialsByCategory(models.CategoryRoofing)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	brands, err := h.store.GetShingleBrands()
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	// Group materials by type for the form dropdowns
	data := map[string]any{
		"Materials":     materials,
		"Brands":        brands,
		"Shingles":      filterByName(materials, "Shingle"),
		"Underlayments": filterByName(materials, "Underlayment"),
		"IceWater":      filterByName(materials, "Ice & Water"),
		"DripEdge":      filterByName(materials, "Drip Edge"),
		"RidgeCap":      filterByNameMulti(materials, []string{"Hip & Ridge", "Ridge Vent"}),
		"Starter":       filterByName(materials, "Starter"),
	}

	h.render(w, "roofing_form.html", data)
}

func (h *Handler) handleRoofingCalculate(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), 400)
		return
	}

	measurements := models.RoofMeasurements{
		CustomerName:    r.FormValue("customer_name"),
		CustomerAddress: r.FormValue("customer_address"),
		CustomerPhone:   r.FormValue("customer_phone"),
		CustomerEmail:   r.FormValue("customer_email"),

		TotalAreaSqFt:   parseFloat(r.FormValue("total_area_sqft")),
		Pitch:           r.FormValue("pitch"),
		RidgeLengthFt:   parseFloat(r.FormValue("ridge_length_ft")),
		HipLengthFt:     parseFloat(r.FormValue("hip_length_ft")),
		ValleyLengthFt:  parseFloat(r.FormValue("valley_length_ft")),
		EaveLengthFt:    parseFloat(r.FormValue("eave_length_ft")),
		RakeLengthFt:    parseFloat(r.FormValue("rake_length_ft")),
		NumPipeBoots:    parseInt(r.FormValue("num_pipe_boots")),
		NumExhaustVents: parseInt(r.FormValue("num_exhaust_vents")),
		LayersToRemove:  parseInt(r.FormValue("layers_to_remove")),
		WasteFactorPct:  parseFloat(r.FormValue("waste_factor_pct")),

		ShingleMaterialID:      parseInt64(r.FormValue("shingle_material_id")),
		UnderlaymentMaterialID: parseInt64(r.FormValue("underlayment_material_id")),
		IceWaterMaterialID:     parseInt64(r.FormValue("ice_water_material_id")),
		DripEdgeMaterialID:     parseInt64(r.FormValue("drip_edge_material_id")),
		RidgeCapMaterialID:     parseInt64(r.FormValue("ridge_cap_material_id")),
		StarterMaterialID:      parseInt64(r.FormValue("starter_material_id")),
	}

	result, err := h.roofCalc.Calculate(measurements)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	// Save the estimate
	estimateJSON, _ := json.Marshal(result)
	id, err := h.store.SaveEstimate(
		result.CustomerName,
		result.CustomerAddress,
		result.ProjectType,
		string(estimateJSON),
	)
	if err != nil {
		log.Printf("Error saving estimate: %v", err)
	}

	data := map[string]any{
		"Result":     result,
		"EstimateID": id,
	}

	// If HTMX request, return partial
	if r.Header.Get("HX-Request") == "true" {
		h.renderPartial(w, "estimate_result.html", data)
		return
	}

	h.render(w, "estimate_result_page.html", data)
}

func (h *Handler) handleProposal(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var estimateJSON string
	err := h.store.DB.QueryRow(
		"SELECT estimate_json FROM saved_estimates WHERE id = ?", id,
	).Scan(&estimateJSON)
	if err != nil {
		http.Error(w, "Estimate not found", 404)
		return
	}

	var result models.EstimateResult
	if err := json.Unmarshal([]byte(estimateJSON), &result); err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	h.renderPartial(w, "proposal.html", map[string]any{
		"Result": result,
	})
}

// --- Material Admin Handlers ---

func (h *Handler) handleMaterialsList(w http.ResponseWriter, r *http.Request) {
	materials, err := h.store.GetAllMaterials()
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	// Group by category
	grouped := map[string][]models.Material{
		"roofing": {},
		"siding":  {},
		"gutters": {},
	}
	for _, m := range materials {
		grouped[string(m.Category)] = append(grouped[string(m.Category)], m)
	}

	h.render(w, "admin_materials.html", map[string]any{
		"Grouped": grouped,
	})
}

func (h *Handler) handleMaterialForm(w http.ResponseWriter, r *http.Request) {
	h.render(w, "admin_material_form.html", map[string]any{
		"Material": models.Material{IsActive: true},
		"IsNew":    true,
	})
}

func (h *Handler) handleMaterialEdit(w http.ResponseWriter, r *http.Request) {
	id := parseInt64(r.PathValue("id"))
	mat, err := h.store.GetMaterial(id)
	if err != nil {
		http.Error(w, "Material not found", 404)
		return
	}
	h.render(w, "admin_material_form.html", map[string]any{
		"Material": mat,
		"IsNew":    false,
	})
}

func (h *Handler) handleMaterialSave(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), 400)
		return
	}

	mat := models.Material{
		ID:              parseInt64(r.FormValue("id")),
		Category:        models.MaterialCategory(r.FormValue("category")),
		Brand:           r.FormValue("brand"),
		ProductLine:     r.FormValue("product_line"),
		Name:            r.FormValue("name"),
		Unit:            r.FormValue("unit"),
		CoveragePerUnit: parseFloat(r.FormValue("coverage_per_unit")),
		CostPerUnit:     parseFloat(r.FormValue("cost_per_unit")),
		PricePerUnit:    parseFloat(r.FormValue("price_per_unit")),
		IsActive:        r.FormValue("is_active") == "on" || r.FormValue("is_active") == "true",
	}

	if _, err := h.store.UpsertMaterial(mat); err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	if r.Header.Get("HX-Request") == "true" {
		w.Header().Set("HX-Redirect", "/admin/materials")
		w.WriteHeader(200)
		return
	}
	http.Redirect(w, r, "/admin/materials", http.StatusSeeOther)
}

func (h *Handler) handleMaterialDelete(w http.ResponseWriter, r *http.Request) {
	id := parseInt64(r.PathValue("id"))
	if err := h.store.DeleteMaterial(id); err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	w.WriteHeader(200)
}

// --- HTMX Partials ---

func (h *Handler) handleShinglesByBrand(w http.ResponseWriter, r *http.Request) {
	brand := r.URL.Query().Get("brand")
	if brand == "" {
		w.Write([]byte(`<option value="">-- Select Brand First --</option>`))
		return
	}

	shingles, err := h.store.GetShinglesByBrand(brand)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(`<option value="">-- Select Shingle --</option>`))
	for _, s := range shingles {
		label := s.ProductLine
		if s.PricePerUnit > 0 {
			label += fmt.Sprintf(" ($%.2f/%s)", s.PricePerUnit, s.Unit)
		}
		fmt.Fprintf(w, `<option value="%d">%s</option>`, s.ID, label)
	}
}

// --- Helpers ---

func (h *Handler) render(w http.ResponseWriter, name string, data any) {
	t, ok := h.templates[name]
	if !ok {
		log.Printf("Template not found: %s", name)
		http.Error(w, "Internal Server Error", 500)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	// Execute the layout wrapper, which calls {{block "content"}} defined by the page
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

func filterByName(materials []models.Material, contains string) []models.Material {
	var result []models.Material
	for _, m := range materials {
		if strings.Contains(m.Name, contains) {
			result = append(result, m)
		}
	}
	return result
}

func filterByNameMulti(materials []models.Material, patterns []string) []models.Material {
	var result []models.Material
	for _, m := range materials {
		for _, p := range patterns {
			if strings.Contains(m.Name, p) {
				result = append(result, m)
				break
			}
		}
	}
	return result
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
