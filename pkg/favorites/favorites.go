package favorites

import (
	"encoding/json"
	"os"
	"sync"
)

// Manager gère le stockage et la manipulation des cartes favorites
type Manager struct {
	filePath string
	mu       sync.RWMutex
	cards    map[string]bool
}

// NewManager crée une nouvelle instance de Manager
func NewManager(filePath string) (*Manager, error) {
	m := &Manager{
		filePath: filePath,
		cards:    make(map[string]bool),
	}

	// Charger les favoris existants
	if err := m.load(); err != nil && !os.IsNotExist(err) {
		return nil, err
	}

	return m, nil
}

// Add ajoute une carte aux favoris
func (m *Manager) Add(cardID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.cards[cardID] = true
	return m.save()
}

// Remove supprime une carte des favoris
func (m *Manager) Remove(cardID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	delete(m.cards, cardID)
	return m.save()
}

// RemoveAll supprime toutes les cartes des favoris
func (m *Manager) RemoveAll() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.cards = make(map[string]bool)
	return m.save()
}

// Contains vérifie si une carte est dans les favoris
func (m *Manager) Contains(cardID string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.cards[cardID]
}

// GetAll retourne la liste de tous les IDs des cartes favorites
func (m *Manager) GetAll() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	cards := make([]string, 0, len(m.cards))
	for cardID := range m.cards {
		cards = append(cards, cardID)
	}
	return cards
}

// load charge les favoris depuis le fichier
func (m *Manager) load() error {
	data, err := os.ReadFile(m.filePath)
	if err != nil {
		return err
	}

	var cards []string
	if err := json.Unmarshal(data, &cards); err != nil {
		return err
	}

	m.cards = make(map[string]bool)
	for _, cardID := range cards {
		m.cards[cardID] = true
	}

	return nil
}

// save sauvegarde les favoris dans le fichier
func (m *Manager) save() error {
	cards := m.GetAll()
	data, err := json.Marshal(cards)
	if err != nil {
		return err
	}

	return os.WriteFile(m.filePath, data, 0644)
}
