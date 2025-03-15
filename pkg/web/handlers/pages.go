package handlers

import (
	"html/template"
	"net/http"

	"cours/pokemon/pkg/services/favorites"
	"cours/pokemon/pkg/services/tcgdex"
)

// PageHandler gère les pages web
type PageHandler struct {
	templates *template.Template
	tcgdex    *tcgdex.Client
	favorites *favorites.Service
}

// NewPageHandler crée un nouveau gestionnaire de pages
func NewPageHandler(templates *template.Template, tcgdex *tcgdex.Client, favorites *favorites.Service) *PageHandler {
	return &PageHandler{
		templates: templates,
		tcgdex:    tcgdex,
		favorites: favorites,
	}
}

// HandleHome gère la page d'accueil
func (h *PageHandler) HandleHome(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	if err := h.templates.ExecuteTemplate(w, "home.html", nil); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// HandleCollection gère la page de collection
func (h *PageHandler) HandleCollection(w http.ResponseWriter, r *http.Request) {
	if err := h.templates.ExecuteTemplate(w, "collection.html", nil); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// HandleFavorites gère la page des favoris
func (h *PageHandler) HandleFavorites(w http.ResponseWriter, r *http.Request) {
	cards := h.favorites.List()
	data := map[string]interface{}{
		"Cards": cards,
	}

	if err := h.templates.ExecuteTemplate(w, "favorites.html", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// HandleSearch gère la page de recherche
func (h *PageHandler) HandleSearch(w http.ResponseWriter, r *http.Request) {
	if err := h.templates.ExecuteTemplate(w, "search.html", nil); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// HandleCategories gère la page des catégories
func (h *PageHandler) HandleCategories(w http.ResponseWriter, r *http.Request) {
	sets, err := h.tcgdex.GetSets()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	data := map[string]interface{}{
		"Sets": sets,
	}

	if err := h.templates.ExecuteTemplate(w, "categories.html", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// HandleAbout gère la page à propos
func (h *PageHandler) HandleAbout(w http.ResponseWriter, r *http.Request) {
	if err := h.templates.ExecuteTemplate(w, "about.html", nil); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
