package handler

import (
	"net/http"
	"strconv"

	"github.com/AugustoGuapo/concretrack-backoffice-be/internal/application"
	"github.com/go-chi/chi/v5"
)

type ReportsHandler struct {
	ReportsService application.ReportsService
}

func NewReportsHandler(service application.ReportsService) *ReportsHandler {
	return &ReportsHandler{ReportsService: service}
}

func (h *ReportsHandler)GenerateReportForOneFamily(w http.ResponseWriter, r *http.Request) {
	projectID := chi.URLParam(r, "ID")
	numericProjectID, err := strconv.Atoi(projectID)
	if err != nil || numericProjectID < 1 {
		http.Error(w, "project ID should be a number greater than zero", http.StatusBadRequest)
	}
	familyID := chi.URLParam(r, "familyID")
	numericFamilyID, err := strconv.Atoi(familyID)
	if err != nil || numericFamilyID < 1 {
		http.Error(w, "family ID should be a number greater than zero", http.StatusBadRequest)
	}
	report, err := h.ReportsService.GenerateReportForOneFamily(numericProjectID, numericFamilyID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	w.Header().Set("Content-Type", "application/pdf")
    w.Header().Set("Content-Disposition", "attachment; filename="+report.Filename)
    w.Write(report.File)
}