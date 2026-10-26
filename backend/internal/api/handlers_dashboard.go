package api

import (
	"net/http"

	"financeapp/backend/internal/domain"
)

func (rt *router) recentMovements(w http.ResponseWriter, r *http.Request) {
	movs, err := rt.movementRepo.ListWithCategory(r.Context())
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "failed to list movements")
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
		out = append(out, movementDTO{Movement: m.Movement, CategoryName: m.CategoryName})
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
		writeJSONError(w, http.StatusInternalServerError, "failed to list movements")
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
	movements, err := rt.movementRepo.ListWithCategory(r.Context())
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "failed to list movements")
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
		if m.EventType != "INCOME" {
			continue
		}
		catName := "Sin categoría"
		if m.CategoryID != "" && m.CategoryName != "" {
			catName = m.CategoryName
		}
		dist[catName] += m.AmountMinor
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
	movements, err := rt.movementRepo.ListWithCategory(r.Context())
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "failed to list movements")
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
		if m.EventType != "EXPENSE" && m.EventType != "DEBIT_PURCHASE" && m.EventType != "CREDIT_PAYMENT" {
			continue
		}
		catName := "Sin categoría"
		if m.CategoryID != "" && m.CategoryName != "" {
			catName = m.CategoryName
		}
		dist[catName] += -m.AmountMinor
	}
	var out []map[string]interface{}
	for k, v := range dist {
		out = append(out, map[string]interface{}{"category": k, "amount_minor": v})
	}
	w.Header().Set("Content-Type", "application/json")
	encodeJSON(w, out)
}
