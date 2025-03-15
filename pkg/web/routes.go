package web

import (
	"html/template"
	"net/http"

	"cours/pokemon/pkg/services/favorites"
	"cours/pokemon/pkg/services/tcgdex"
	"cours/pokemon/pkg/web/handlers"
)

// RegisterRoutes enregistre les routes web
func RegisterRoutes(mux *http.ServeMux, templates *template.Template, tcgdex *tcgdex.Client, favorites *favorites.Service) {
	pageHandler := handlers.NewPageHandler(templates, tcgdex, favorites)

	// Routes des pages
	mux.HandleFunc("/", pageHandler.HandleHome)
	mux.HandleFunc("/collection", pageHandler.HandleCollection)
	mux.HandleFunc("/favorites", pageHandler.HandleFavorites)
	mux.HandleFunc("/search", pageHandler.HandleSearch)
	mux.HandleFunc("/categories", pageHandler.HandleCategories)
	mux.HandleFunc("/about", pageHandler.HandleAbout)
}
