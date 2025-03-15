package server

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"path/filepath"

	"cours/pokemon/pkg/api"
	"cours/pokemon/pkg/services/favorites"
	"cours/pokemon/pkg/services/tcgdex"
	"cours/pokemon/pkg/web"
)

// Server représente le serveur HTTP
type Server struct {
	tcgdex    *tcgdex.Client
	favorites *favorites.Service
	templates *template.Template
	staticDir string
	mux       *http.ServeMux
}

// NewServer crée une nouvelle instance du serveur
func NewServer(tcgdex *tcgdex.Client, favorites *favorites.Service, templatesDir, staticDir string) (*Server, error) {
	templates, err := template.ParseGlob(filepath.Join(templatesDir, "*.html"))
	if err != nil {
		return nil, fmt.Errorf("erreur lors du chargement des templates: %w", err)
	}

	return &Server{
		tcgdex:    tcgdex,
		favorites: favorites,
		templates: templates,
		staticDir: staticDir,
		mux:       http.NewServeMux(),
	}, nil
}

// Start démarre le serveur HTTP
func (s *Server) Start(addr string) error {
	// Routes statiques
	fs := http.FileServer(http.Dir(s.staticDir))
	s.mux.Handle("/static/", http.StripPrefix("/static/", fs))

	// Enregistre les routes
	api.RegisterRoutes(s.mux, s.tcgdex, s.favorites)
	web.RegisterRoutes(s.mux, s.templates, s.tcgdex, s.favorites)

	log.Printf("Serveur démarré sur %s", addr)
	return http.ListenAndServe(addr, s.mux)
}
