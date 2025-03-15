package cache

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"sync"
	"time"

	"cours/pokemon/pkg/handlers/sets"
	"cours/pokemon/pkg/models"
)

const (
	cacheDuration = 24 * time.Hour
	apiBaseURL    = "https://api.tcgdex.net/v2/fr"
)

// TCGCard représente les données brutes d'une carte de l'API
type TCGCard struct {
	ID    string   `json:"id"`
	Name  string   `json:"name"`
	Types []string `json:"types"`
	Set   struct {
		Name string `json:"name"`
	} `json:"set"`
}

// Cache gère le cache des cartes
type Cache struct {
	sync.RWMutex
	names       map[string]string    // Cache des noms de cartes
	images      map[string][]byte    // Cache des images
	lastUpdated map[string]time.Time // Horodatage de dernière mise à jour
}

var instance = &Cache{
	names:       make(map[string]string),
	images:      make(map[string][]byte),
	lastUpdated: make(map[string]time.Time),
}

// GetInstance retourne l'instance unique du cache
func GetInstance() *Cache {
	return instance
}

// GetCardInfo récupère les informations d'une carte depuis le cache ou l'API
func (c *Cache) GetCardInfo(series string, number int) models.Card {
	paddedNumber := fmt.Sprintf("%d", number)
	if number < 10 {
		paddedNumber = fmt.Sprintf("00%d", number)
	} else if number < 100 {
		paddedNumber = fmt.Sprintf("0%d", number)
	}

	cardID := fmt.Sprintf("%s-%s", series, paddedNumber)

	c.RLock()
	if name, exists := c.names[cardID]; exists {
		if time.Since(c.lastUpdated[cardID]) < cacheDuration {
			c.RUnlock()
			return models.Card{
				ID:   cardID,
				Name: name,
				Set: models.Set{
					ID:   series,
					Name: sets.GetSetName(series),
				},
			}
		}
	}
	c.RUnlock()

	urls := []string{
		fmt.Sprintf("%s/cards/%s", apiBaseURL, cardID),
		fmt.Sprintf("%s/cards/%s-%d", apiBaseURL, series, number),
	}

	var tcgCard TCGCard
	var err error
	var resp *http.Response

	for _, url := range urls {
		log.Printf("Tentative avec URL: %s", url)
		resp, err = http.Get(url)
		if err == nil && resp.StatusCode == http.StatusOK {
			defer resp.Body.Close()
			if err := json.NewDecoder(resp.Body).Decode(&tcgCard); err == nil && tcgCard.Name != "" {
				log.Printf("Carte trouvée: %s, Types: %v", tcgCard.Name, tcgCard.Types)
				break
			}
		}
		if resp != nil {
			resp.Body.Close()
		}
	}

	if tcgCard.Name == "" {
		log.Printf("Aucune donnée trouvée pour la carte %s", cardID)
		return models.Card{
			ID:   cardID,
			Name: fmt.Sprintf("Carte %s #%d", series, number),
			Set: models.Set{
				ID:   series,
				Name: sets.GetSetName(series),
			},
		}
	}

	c.Lock()
	c.names[cardID] = tcgCard.Name
	c.lastUpdated[cardID] = time.Now()
	c.Unlock()

	return models.Card{
		ID:   cardID,
		Name: tcgCard.Name,
		Set: models.Set{
			ID:   series,
			Name: sets.GetSetName(series),
		},
		Types: tcgCard.Types,
	}
}

// GetCardImage récupère une image depuis le cache ou l'API
func (c *Cache) GetCardImage(series, number, quality string) ([]byte, error) {
	imageKey := fmt.Sprintf("%s-%s-%s", series, number, quality)

	c.RLock()
	if img, exists := c.images[imageKey]; exists {
		if time.Since(c.lastUpdated[imageKey]) < cacheDuration {
			c.RUnlock()
			return img, nil
		}
	}
	c.RUnlock()

	// Construire les différents formats d'URL possibles
	urls := make([]string, 0, 4)

	// Format avec padding (097)
	paddedNumber := number
	if len(number) == 1 {
		paddedNumber = "00" + number
	} else if len(number) == 2 {
		paddedNumber = "0" + number
	}

	// Formats d'URL à essayer
	urls = append(urls,
		fmt.Sprintf("https://assets.tcgdex.net/fr/swsh/%s/%s/%s.webp", series, paddedNumber, quality),
		fmt.Sprintf("https://assets.tcgdex.net/fr/swsh/%s/%s/%s.png", series, paddedNumber, quality),
		fmt.Sprintf("https://assets.tcgdex.net/fr/%s/%s/%s.webp", series, paddedNumber, quality),
		fmt.Sprintf("https://assets.tcgdex.net/fr/%s/%s/%s.png", series, paddedNumber, quality))

	var lastErr error
	for _, url := range urls {
		log.Printf("Tentative avec URL: %s", url)
		img, err := fetchImage(url)
		if err == nil {
			c.Lock()
			c.images[imageKey] = img
			c.lastUpdated[imageKey] = time.Now()
			c.Unlock()
			return img, nil
		}
		lastErr = err
	}

	// Si la qualité demandée est "high", essayer automatiquement en "low"
	if quality == "high" {
		return c.GetCardImage(series, number, "low")
	}

	return nil, fmt.Errorf("aucune image trouvée pour %s/%s (dernière erreur: %v)", series, number, lastErr)
}

// fetchImage récupère une image depuis une URL
func fetchImage(url string) ([]byte, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36")
	req.Header.Set("Accept", "image/webp,image/apng,image/*,*/*;q=0.8")
	req.Header.Set("Accept-Language", "fr-FR,fr;q=0.9,en-US;q=0.8,en;q=0.7")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("status code: %d", resp.StatusCode)
	}

	return io.ReadAll(resp.Body)
}
