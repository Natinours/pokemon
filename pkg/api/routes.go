package api

import (
	"net/http"

	"cours/pokemon/pkg/api/handlers"
	"cours/pokemon/pkg/services/favorites"
	"cours/pokemon/pkg/services/tcgdex"
)

// RegisterRoutes enregistre les routes de l'API
func RegisterRoutes(mux *http.ServeMux, tcgdex *tcgdex.Client, favorites *favorites.Service) {
	cardsHandler := handlers.NewCardsHandler(tcgdex)
	favoritesHandler := handlers.NewFavoritesHandler(favorites, tcgdex)

	// Routes de l'API
	mux.HandleFunc("/api/cards/search", cardsHandler.HandleSearch)
	mux.HandleFunc("/api/cards/", cardsHandler.HandleGetCard)
	mux.HandleFunc("/api/favorites/", favoritesHandler.HandleFavoriteCard)
}
