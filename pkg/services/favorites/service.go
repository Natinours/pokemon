package favorites

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"cours/pokemon/pkg/models"
)

// Service gère les cartes favorites
type Service struct {
	filePath string
	mu       sync.RWMutex
	cards    map[string]*models.Card
}

// NewService crée une nouvelle instance du service de favoris
func NewService(dataDir string) (*Service, error) {
	filePath := filepath.Join(dataDir, "favorites.json")

	service := &Service{
		filePath: filePath,
		cards:    make(map[string]*models.Card),
	}

	// Charge les favoris existants
	if err := service.load(); err != nil && !os.IsNotExist(err) {
		return nil, fmt.Errorf("erreur lors du chargement des favoris: %w", err)
	}

	return service, nil
}

// Add ajoute une carte aux favoris
func (s *Service) Add(card *models.Card) error {
	if !card.IsValid() {
		return fmt.Errorf("carte invalide")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	s.cards[card.ID] = card
	return s.save()
}

// Remove supprime une carte des favoris
func (s *Service) Remove(cardID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.cards, cardID)
	return s.save()
}

// Get récupère une carte favorite par son ID
func (s *Service) Get(cardID string) (*models.Card, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	card, exists := s.cards[cardID]
	return card, exists
}

// List retourne la liste de toutes les cartes favorites
func (s *Service) List() []*models.Card {
	s.mu.RLock()
	defer s.mu.RUnlock()

	cards := make([]*models.Card, 0, len(s.cards))
	for _, card := range s.cards {
		cards = append(cards, card)
	}
	return cards
}

// Clear supprime toutes les cartes favorites
func (s *Service) Clear() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.cards = make(map[string]*models.Card)
	return s.save()
}

// load charge les favoris depuis le fichier
func (s *Service) load() error {
	data, err := os.ReadFile(s.filePath)
	if err != nil {
		return err
	}

	var cards []*models.Card
	if err := json.Unmarshal(data, &cards); err != nil {
		return fmt.Errorf("erreur lors du décodage des favoris: %w", err)
	}

	s.cards = make(map[string]*models.Card, len(cards))
	for _, card := range cards {
		s.cards[card.ID] = card
	}

	return nil
}

// save sauvegarde les favoris dans le fichier
func (s *Service) save() error {
	cards := make([]*models.Card, 0, len(s.cards))
	for _, card := range s.cards {
		cards = append(cards, card)
	}

	data, err := json.MarshalIndent(cards, "", "  ")
	if err != nil {
		return fmt.Errorf("erreur lors de l'encodage des favoris: %w", err)
	}

	if err := os.MkdirAll(filepath.Dir(s.filePath), 0755); err != nil {
		return fmt.Errorf("erreur lors de la création du dossier: %w", err)
	}

	if err := os.WriteFile(s.filePath, data, 0644); err != nil {
		return fmt.Errorf("erreur lors de la sauvegarde des favoris: %w", err)
	}

	return nil
}
