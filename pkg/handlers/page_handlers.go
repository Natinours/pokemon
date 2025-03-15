package handlers

import (
	"cours/pokemon/pkg/models"
	"html/template"
	"log"
	"net/http"
	"path/filepath"
)

// Structure pour les données de la page catégories
type CategoriesPageData struct {
	SeriesList  []models.PokemonSeries
	TotalSeries int
}

// Fonction helper pour obtenir le nombre de cartes dans un set
func getSetCardCount(setID string) int {
	if seriesInfo, ok := models.GetSeriesByID(setID); ok {
		return seriesInfo.CardCount
	}
	return 0
}

// HandleHome gère la page d'accueil
func HandleHome(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.ParseFiles(
		filepath.Join("templates", "layout.html"),
		filepath.Join("templates", "home.html"),
	))

	if err := tmpl.ExecuteTemplate(w, "layout", nil); err != nil {
		log.Printf("Erreur lors de l'exécution du template : %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// HandleCategories gère la page des catégories
func HandleCategories(w http.ResponseWriter, r *http.Request) {
	// Récupérer toutes les séries
	seriesList := models.GetAllSeries()

	// Créer le template avec les fonctions
	tmpl := template.Must(template.New("layout.html").ParseFiles(
		filepath.Join("templates", "layout.html"),
		filepath.Join("templates", "categories.html"),
	))

	// Préparer les données
	data := struct {
		SeriesList []models.PokemonSeries
	}{
		SeriesList: seriesList,
	}

	// Exécuter le template
	if err := tmpl.ExecuteTemplate(w, "layout", data); err != nil {
		log.Printf("Erreur lors de l'exécution du template : %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// HandleAbout gère la page À propos
func HandleAbout(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.ParseFiles(
		filepath.Join("templates", "layout.html"),
		filepath.Join("templates", "about.html"),
	))

	if err := tmpl.ExecuteTemplate(w, "layout", nil); err != nil {
		log.Printf("Erreur lors de l'exécution du template : %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
