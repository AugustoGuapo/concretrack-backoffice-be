package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/AugustoGuapo/concretrack-backoffice-be/internal/domain/project"
	"github.com/go-chi/chi/v5"
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

func (h *ProjectHandler) SaveProject(w http.ResponseWriter, r *http.Request) {
	project := &project.Project{}
	if err := json.NewDecoder(r.Body).Decode(project); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	createdProject, err := h.service.SaveProject(project)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(createdProject)
}

func(h *ProjectHandler) GetProjectsByClientID(w http.ResponseWriter, r *http.Request) {
	clientID, err := strconv.Atoi(chi.URLParam(r, "clientID"))

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}	

	projects, err := h.service.GetProjectsByClientID(clientID)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(projects)

}
