package main

import (
	"flag"
	"log"
	"os"
	"path/filepath"

	"cours/pokemon/pkg/server"
	"cours/pokemon/pkg/services/favorites"
	"cours/pokemon/pkg/services/tcgdex"
)

func main() {
	// Configuration par ligne de commande
	addr := flag.String("addr", ":8080", "Adresse du serveur HTTP")
	dataDir := flag.String("data", "data", "Dossier des données")
	templatesDir := flag.String("templates", "templates", "Dossier des templates")
	staticDir := flag.String("static", "static", "Dossier des fichiers statiques")
	flag.Parse()

	// Crée les dossiers nécessaires
	dirs := []string{*dataDir, *templatesDir, *staticDir}
	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			log.Fatalf("Erreur lors de la création du dossier %s: %v", dir, err)
		}
	}

	// Initialise les services
	tcgdexClient := tcgdex.NewClient()

	favoritesService, err := favorites.NewService(*dataDir)
	if err != nil {
		log.Fatalf("Erreur lors de l'initialisation du service des favoris: %v", err)
	}

	// Initialise le serveur HTTP
	srv, err := server.NewServer(tcgdexClient, favoritesService, *templatesDir, *staticDir)
	if err != nil {
		log.Fatalf("Erreur lors de l'initialisation du serveur: %v", err)
	}

	// Obtient les chemins absolus
	dataDirAbs, err := filepath.Abs(*dataDir)
	if err != nil {
		log.Printf("Attention: impossible d'obtenir le chemin absolu pour %s: %v", *dataDir, err)
		dataDirAbs = *dataDir
	}

	templatesDirAbs, err := filepath.Abs(*templatesDir)
	if err != nil {
		log.Printf("Attention: impossible d'obtenir le chemin absolu pour %s: %v", *templatesDir, err)
		templatesDirAbs = *templatesDir
	}

	staticDirAbs, err := filepath.Abs(*staticDir)
	if err != nil {
		log.Printf("Attention: impossible d'obtenir le chemin absolu pour %s: %v", *staticDir, err)
		staticDirAbs = *staticDir
	}

	// Démarre le serveur
	log.Printf("Démarrage du serveur sur %s", *addr)
	log.Printf("Dossier des données: %s", dataDirAbs)
	log.Printf("Dossier des templates: %s", templatesDirAbs)
	log.Printf("Dossier des fichiers statiques: %s", staticDirAbs)

	if err := srv.Start(*addr); err != nil {
		log.Fatalf("Erreur lors du démarrage du serveur: %v", err)
	}
}
