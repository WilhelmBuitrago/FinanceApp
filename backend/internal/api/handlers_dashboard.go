package api

import (
	"net/http"

	"financeapp/backend/internal/domain"
)

func (rt *router) recentMovements(w http.ResponseWriter, r *http.Request) {
	movs, err := rt.movementRepo.List(r.Context())
	if err != nil {
		http.Error(w, `{"error":"failed to list movements"}`, http.StatusInternalServerError)
		return
	}
	if len(movs) > 5 {
		movs = movs[:5]
	}
	type movementDTO struct {
		domain.Movement
		CategoryName string `json:"categoryName"`
	}
	out := make([]movementDTO, 0, len(movs))
	for _, m := range movs {
		fe, _ := rt.eventRepo.GetByID(r.Context(), m.FinancialEventID)
		catName := ""
		if fe.CategoryID != nil {
			if cat, err := rt.categoryRepo.GetByID(r.Context(), *fe.CategoryID); err == nil {
				catName = cat.Name
			}
		}
		out = append(out, movementDTO{Movement: m, CategoryName: catName})
	}
	w.Header().Set("Content-Type", "application/json")
	encodeJSON(w, out)
}

// cashFlow aggregates movements per day (in-memory aggregation on purpose:
// the dataset is local and small).
func (rt *router) cashFlow(w http.ResponseWriter, r *http.Request) {
	from := r.URL.Query().Get("from")
	to := r.URL.Query().Get("to")
	movements, err := rt.movementRepo.List(r.Context())
	if err != nil {
		http.Error(w, `{"error":"failed to list movements"}`, http.StatusInternalServerError)
		return
	}
	agg := make(map[string]struct{ inflow, outflow int64 })
	for _, m := range movements {
		if from != "" && m.MovementDate < from {
			continue
		}
		if to != "" && m.MovementDate > to {
			continue
		}
		// Skip cancelled
		if m.Status != "ACTIVE" {
			continue
		}
		date := m.MovementDate
		cur := agg[date]
		if m.AmountMinor > 0 {
			cur.inflow += m.AmountMinor
		} else {
			cur.outflow += -m.AmountMinor
		}
		agg[date] = cur
	}
	type point struct {
		Date    string `json:"date"`
		Inflow  int64  `json:"inflow"`
		Outflow int64  `json:"outflow"`
		Net     int64  `json:"net"`
	}
	var out []point
	for d, v := range agg {
		out = append(out, point{Date: d, Inflow: v.inflow, Outflow: v.outflow, Net: v.inflow - v.outflow})
	}
	w.Header().Set("Content-Type", "application/json")
	encodeJSON(w, out)
}

func (rt *router) incomeDistribution(w http.ResponseWriter, r *http.Request) {
	from := r.URL.Query().Get("from")
	to := r.URL.Query().Get("to")
	movements, err := rt.movementRepo.List(r.Context())
	if err != nil {
		http.Error(w, `{"error":"failed to list movements"}`, http.StatusInternalServerError)
		return
	}
	dist := make(map[string]int64)
	catNames := make(map[string]string)
	for _, m := range movements {
		if from != "" && m.MovementDate < from {
			continue
		}
		if to != "" && m.MovementDate > to {
			continue
		}
		if m.Status != "ACTIVE" {
			continue
		}
		if m.FinancialEventID == "" {
			continue
		}
		evt, err := rt.eventRepo.GetByID(r.Context(), m.FinancialEventID)
		if err != nil {
			continue
		}
		if evt.EventType != "INCOME" {
			continue
		}
		var catName string
		if evt.CategoryID == nil || *evt.CategoryID == "" {
			catName = "Sin categoría"
		} else {
			catID := *evt.CategoryID
			cat, err := rt.categoryRepo.GetByID(r.Context(), catID)
			if err != nil {
				catName = "Sin categoría"
			} else {
				catName = cat.Name
			}
		}
		key := catName
		dist[key] += m.AmountMinor
		catNames[key] = catName
	}
	var out []map[string]interface{}
	for k, v := range dist {
		out = append(out, map[string]interface{}{"category": k, "amount_minor": v})
	}
	w.Header().Set("Content-Type", "application/json")
	encodeJSON(w, out)
}

func (rt *router) expenseDistribution(w http.ResponseWriter, r *http.Request) {
	from := r.URL.Query().Get("from")
	to := r.URL.Query().Get("to")
	movements, err := rt.movementRepo.List(r.Context())
	if err != nil {
		http.Error(w, `{"error":"failed to list movements"}`, http.StatusInternalServerError)
		return
	}
	dist := make(map[string]int64)
	for _, m := range movements {
		if from != "" && m.MovementDate < from {
			continue
		}
		if to != "" && m.MovementDate > to {
			continue
		}
		if m.Status != "ACTIVE" {
			continue
		}
		if m.FinancialEventID == "" {
			continue
		}
		evt, err := rt.eventRepo.GetByID(r.Context(), m.FinancialEventID)
		if err != nil {
			continue
		}
		if evt.EventType != "EXPENSE" && evt.EventType != "DEBIT_PURCHASE" && evt.EventType != "CREDIT_PAYMENT" {
			continue
		}
		var catName string
		if evt.CategoryID == nil || *evt.CategoryID == "" {
			catName = "Sin categoría"
		} else {
			catID := *evt.CategoryID
			cat, err := rt.categoryRepo.GetByID(r.Context(), catID)
			if err != nil {
				catName = "Sin categoría"
			} else {
				catName = cat.Name
			}
		}
		key := catName
		dist[key] += -m.AmountMinor
	}
	var out []map[string]interface{}
	for k, v := range dist {
		out = append(out, map[string]interface{}{"category": k, "amount_minor": v})
	}
	w.Header().Set("Content-Type", "application/json")
	encodeJSON(w, out)
}
