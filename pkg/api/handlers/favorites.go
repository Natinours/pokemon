package handlers

import (
	"net/http"
	"strings"

	"cours/pokemon/pkg/services/favorites"
	"cours/pokemon/pkg/services/tcgdex"
)

// FavoritesHandler gère les requêtes API liées aux favoris
type FavoritesHandler struct {
	favorites *favorites.Service
	tcgdex    *tcgdex.Client
}

// NewFavoritesHandler crée un nouveau gestionnaire de favoris
func NewFavoritesHandler(favorites *favorites.Service, tcgdex *tcgdex.Client) *FavoritesHandler {
	return &FavoritesHandler{
		favorites: favorites,
		tcgdex:    tcgdex,
	}
}

// HandleFavoriteCard gère les opérations sur une carte favorite
func (h *FavoritesHandler) HandleFavoriteCard(w http.ResponseWriter, r *http.Request) {
	// Extrait l'ID de la carte de l'URL
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 3 {
		http.Error(w, "ID de carte manquant", http.StatusBadRequest)
		return
	}
	cardID := parts[len(parts)-1]

	switch r.Method {
	case http.MethodPost:
		// Ajoute aux favoris
		card, err := h.tcgdex.GetCard(cardID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if err := h.favorites.Add(card); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusCreated)

	case http.MethodDelete:
		// Supprime des favoris
		if err := h.favorites.Remove(cardID); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusNoContent)

	default:
		http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
	}
}
