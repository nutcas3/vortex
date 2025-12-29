package handlers

import (
	"encoding/json"
	"net/http"
)

type AccountHandler struct {
	// Add dependencies here (account repository, etc.)
}

func NewAccountHandler() *AccountHandler {
	return &AccountHandler{}
}

// Account handles account information
func (h *AccountHandler) Account(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{"message": "GET account - TODO: Implement"})
}

// Positions handles position management
func (h *AccountHandler) Positions(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	switch r.Method {
	case http.MethodGet:
		// Get positions
		json.NewEncoder(w).Encode(map[string]string{"message": "GET positions - TODO: Implement"})
	case http.MethodPost:
		// Create/modify position
		json.NewEncoder(w).Encode(map[string]string{"message": "POST position - TODO: Implement"})
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}
