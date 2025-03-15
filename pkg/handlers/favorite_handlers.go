package handlers

import (
	"cours/pokemon/pkg/favorites"
	"encoding/json"
	"html/template"
	"log"
	"net/http"
	"path/filepath"
)

var favManager *favorites.Manager

func init() {
	var err error
	favManager, err = favorites.NewManager("data/favorites.json")
	if err != nil {
		log.Fatal("Erreur lors de l'initialisation du gestionnaire de favoris:", err)
	}
}

// HandleFavorites gère l'affichage des favoris
func HandleFavorites(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles(
		filepath.Join("templates", "layout.html"),
		filepath.Join("templates", "favorites.html"),
	)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Pas besoin de passer de données car nous utilisons maintenant le localStorage
	tmpl.ExecuteTemplate(w, "layout", nil)
}

// HandleAddFavorite gère l'ajout d'une carte aux favoris
func HandleAddFavorite(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
		return
	}

	var data struct {
		CardID string `json:"cardId"`
	}

	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		http.Error(w, "Données invalides", http.StatusBadRequest)
		return
	}

	if err := favManager.Add(data.CardID); err != nil {
		http.Error(w, "Erreur lors de l'ajout aux favoris", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}

// HandleRemoveFavorite gère la suppression d'une carte des favoris
func HandleRemoveFavorite(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
		return
	}

	var data struct {
		CardID string `json:"cardId"`
	}

	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		http.Error(w, "Données invalides", http.StatusBadRequest)
		return
	}

	if err := favManager.Remove(data.CardID); err != nil {
		http.Error(w, "Erreur lors de la suppression des favoris", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}

// HandleRemoveAllFavorites gère la suppression de tous les favoris
func HandleRemoveAllFavorites(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
		return
	}

	if err := favManager.RemoveAll(); err != nil {
		http.Error(w, "Erreur lors de la suppression des favoris", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}
