package api

import (
	"net/http"
)

// Read-only traceability chain: Movement -> Financial Event -> Source Entry -> Document

func (rt *router) listDocuments(w http.ResponseWriter, r *http.Request) {
	documents, err := rt.docRepo.List(r.Context())
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "failed to list documents")
		return
	}
	w.Header().Set("Content-Type", "application/json")
	encodeJSON(w, documents)
}

func (rt *router) recentDocuments(w http.ResponseWriter, r *http.Request) {
	docs, err := rt.docRepo.List(r.Context())
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "failed to list documents")
		return
	}
	// limit 3
	if len(docs) > 3 {
		docs = docs[:3]
	}
	w.Header().Set("Content-Type", "application/json")
	encodeJSON(w, docs)
}

func (rt *router) getDocument(w http.ResponseWriter, r *http.Request) {
	doc, err := rt.docRepo.GetByID(r.Context(), r.PathValue("id"))
	if err != nil {
		writeLookupError(w, err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	encodeJSON(w, doc)
}

func (rt *router) getFinancialEvent(w http.ResponseWriter, r *http.Request) {
	event, err := rt.eventRepo.GetByID(r.Context(), r.PathValue("id"))
	if err != nil {
		writeLookupError(w, err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	encodeJSON(w, event)
}

func (rt *router) getSourceEntry(w http.ResponseWriter, r *http.Request) {
	entry, err := rt.sourceEntryRepo.GetByID(r.Context(), r.PathValue("id"))
	if err != nil {
		writeLookupError(w, err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	encodeJSON(w, entry)
}
