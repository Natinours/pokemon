package handlers

import (
	"encoding/json"
	"net/http"
	"strings"

	"cours/pokemon/pkg/models"
	"cours/pokemon/pkg/services/tcgdex"
)

// CardsHandler gère les requêtes API liées aux cartes
type CardsHandler struct {
	tcgdex *tcgdex.Client
}

// NewCardsHandler crée un nouveau gestionnaire de cartes
func NewCardsHandler(tcgdex *tcgdex.Client) *CardsHandler {
	return &CardsHandler{
		tcgdex: tcgdex,
	}
}

// HandleSearch gère la recherche de cartes
func (h *CardsHandler) HandleSearch(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
		return
	}

	options := models.SearchRequest{
		Query:    r.URL.Query().Get("q"),
		Series:   r.URL.Query().Get("series"),
		Page:     1, // TODO: pagination
		PageSize: 20,
	}

	cards, err := h.tcgdex.SearchCards(options)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(cards)
}

// HandleGetCard gère la récupération d'une carte
func (h *CardsHandler) HandleGetCard(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
		return
	}

	// Extrait l'ID de la carte de l'URL
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 3 {
		http.Error(w, "ID de carte manquant", http.StatusBadRequest)
		return
	}
	cardID := parts[len(parts)-1]

	card, err := h.tcgdex.GetCard(cardID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(card)
}
