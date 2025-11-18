package handler

import (
	"encoding/json"
	"net/http"

	"github.com/AugustoGuapo/concretrack-backoffice-be/internal/domain/project"
)

type ProjectHandler struct {
	service *project.Service
}

func NewProjectHandler(service *project.Service) *ProjectHandler {
	return &ProjectHandler{service: service}
}

func (h *ProjectHandler) GetProjectByID(w http.ResponseWriter, r *http.Request, ID int) {
    project, err := h.service.GetProjectByID(ID)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(project)
}

func (h *ProjectHandler) GetProjects(w http.ResponseWriter, r *http.Request, page int) {
	projects, err := h.service.GetProjects(page)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(projects)
}
