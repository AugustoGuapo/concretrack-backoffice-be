package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/AugustoGuapo/concretrack-backoffice-be/internal/domain/client"
	"github.com/go-chi/chi/v5"
)

type ClientHandler struct {
	service *client.Service
}

func NewClientHandler(service *client.Service) *ClientHandler {
	return &ClientHandler{service: service}
}

func (h *ClientHandler) GetClient(w http.ResponseWriter, r *http.Request) {
	clientID, err := strconv.Atoi(chi.URLParam(r, "ID"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	client, err := h.service.GetClient(clientID)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(client)
}

func (h *ClientHandler) GetAllClients(w http.ResponseWriter, r *http.Request) {

	clients, err := h.service.GetAllClients()

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(clients)
}

func (h *ClientHandler) SaveClient(w http.ResponseWriter, r *http.Request) {
	client := &client.Client{}

	if err := json.NewDecoder(r.Body).Decode(client); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	createdClient, err := h.service.SaveClient(client)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(createdClient)
}
