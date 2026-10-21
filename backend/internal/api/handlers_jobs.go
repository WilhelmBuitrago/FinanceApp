package api

import (
	"net/http"
	"os"
	"path/filepath"
)

func (rt *router) getJobByPath(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Path[len("/jobs/"):]
	GetJobHandler(w, r, rt.importRepo, id)
}

// jobParsed and jobManifest read the artifacts the processor left on disk
// under the configured artifacts path.
func (rt *router) jobParsed(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	// Load job to get processing_id
	job, err := rt.importRepo.GetByID(r.Context(), id)
	if err != nil {
		http.Error(w, `{"error":"not found"}`, http.StatusNotFound)
		return
	}
	if job.ProcessingID == nil {
		http.Error(w, `{"error":"processing not ready"}`, http.StatusBadRequest)
		return
	}
	artifactPath := filepath.Join(rt.cfg.ArtifactsPath, *job.ProcessingID, "parsed.json")
	data, err := os.ReadFile(artifactPath)
	if err != nil {
		http.Error(w, `{"error":"parsed not found"}`, http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(data)
}

func (rt *router) jobManifest(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	job, err := rt.importRepo.GetByID(r.Context(), id)
	if err != nil {
		http.Error(w, `{"error":"not found"}`, http.StatusNotFound)
		return
	}
	if job.ProcessingID == nil {
		http.Error(w, `{"error":"processing not ready"}`, http.StatusBadRequest)
		return
	}
	artifactPath := filepath.Join(rt.cfg.ArtifactsPath, *job.ProcessingID, "manifest.json")
	data, err := os.ReadFile(artifactPath)
	if err != nil {
		http.Error(w, `{"error":"manifest not found"}`, http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(data)
}

// confirmJob runs the injected import service and reports the final job
// status.
func (rt *router) confirmJob(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	job, err := rt.importRepo.GetByID(r.Context(), id)
	if err != nil {
		http.Error(w, `{"error":"not found"}`, http.StatusNotFound)
		return
	}
	if job.Status == "SUCCESS" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"SUCCESS","id":"` + id + `"}`))
		return
	}
	if job.Status == "REVIEW" {
		// check pending decisions - simplified: reject if status is REVIEW
		http.Error(w, `{"error":"decisions pending"}`, http.StatusConflict)
		return
	}
	// Call real import service via injected handler
	if rt.confirmHandler != nil {
		if err := rt.confirmHandler(r.Context(), id); err != nil {
			http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusInternalServerError)
			return
		}
	}
	// reload job
	job, err = rt.importRepo.GetByID(r.Context(), id)
	if err != nil {
		http.Error(w, `{"error":"job reload failed"}`, http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"` + job.Status + `","id":"` + id + `"}`))
}
