package handler

import (
	"encoding/json"
	"net/http"

	"github.com/AugustoGuapo/concretrack-backoffice-be/internal/domain/member"
)

type MemberHandler struct {
	service *member.Service
}

func NewMemberHandler(service *member.Service) *MemberHandler {
	return &MemberHandler{service: service}
}

func (h *MemberHandler) SaveMembers(w http.ResponseWriter, r *http.Request) {
	var members []*member.Member

	if err := json.NewDecoder(r.Body).Decode(&members); err != nil {
		http.Error(w, "invalid JSON body: "+err.Error(), http.StatusBadRequest)
		return
	}

	saved, err := h.service.SaveMembers(members)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(saved)
}
