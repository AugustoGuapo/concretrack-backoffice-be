package handler

import (
	"encoding/json"
	"net/http"

	"github.com/AugustoGuapo/concretrack-backoffice-be/internal/domain/family"
)

type FamilyHandler struct {
	service *family.Service
}

func NewFamilyHandler(service *family.Service) *FamilyHandler {
	return &FamilyHandler{service: service}
}

func (h *FamilyHandler) SaveFamily(w http.ResponseWriter, r *http.Request) {
	family := &family.Family{}
	if err := json.NewDecoder(r.Body).Decode(family); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	createdFamily, err := h.service.SaveFamily(*family)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(createdFamily)
}
