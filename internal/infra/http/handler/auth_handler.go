package handler

import (
	"encoding/json"
	"net/http"

	"github.com/AugustoGuapo/concretrack-backoffice-be/internal/domain/user"
)

type AuthHandler struct {
	service *user.Service
}

func NewAuthHandler(service *user.Service) *AuthHandler {
	return &AuthHandler{service: service}
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var creds struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&creds); err != nil {
		http.Error(w, "Invalid JSON body", http.StatusBadRequest)
		return
	}
	u, loginErr := h.service.Login(creds.Username, creds.Password)
	if loginErr != nil {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	response := struct {
		ID        int    `json:"id"`
		Username  string `json:"username"`
		FirstName string `json:"firstName"`
		LastName  string `json:"lastName"`
	}{
		ID:        u.ID,
		Username:  u.Username,
		FirstName: u.FirstName,
		LastName:  u.LastName,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
