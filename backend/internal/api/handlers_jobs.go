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
	_, _ = w.Write(data)
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
	_, _ = w.Write(data)
}
