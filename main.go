package main

import (
	"cours/pokemon/pkg/handlers"
	"log"
	"net/http"
	"os"
)

func main() {
	// Création du dossier data s'il n'existe pas
	if err := os.MkdirAll("data", 0755); err != nil {
		log.Fatal("Erreur lors de la création du dossier data:", err)
	}

	// Configuration des routes
	setupRoutes()

	// Démarrage du serveur
	log.Println("Serveur démarré sur http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func setupRoutes() {
	// Routes statiques
	fs := http.FileServer(http.Dir("static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	// Routes principales
	http.HandleFunc("/", handlers.HandleHome)
	http.HandleFunc("/collection", handlers.HandleCollection)
	http.HandleFunc("/card/", handlers.HandleCardDetails)
	http.HandleFunc("/search", handlers.HandleSearch)
	http.HandleFunc("/categories", handlers.HandleCategories)
	http.HandleFunc("/favorites", handlers.HandleFavorites)
	http.HandleFunc("/about", handlers.HandleAbout)

	// Routes API pour les favoris
	http.HandleFunc("/api/favorites/add", handlers.HandleAddFavorite)
	http.HandleFunc("/api/favorites/remove", handlers.HandleRemoveFavorite)
	http.HandleFunc("/api/favorites/remove-all", handlers.HandleRemoveAllFavorites)

	// Route pour le proxy des images
	http.HandleFunc("/proxy/card-image/", handlers.HandleCardImageProxy)
}
