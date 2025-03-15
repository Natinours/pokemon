package tcgdex

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"cours/pokemon/pkg/models"
)

const (
	baseURL        = "https://api.tcgdex.net/v2/fr"
	defaultTimeout = 10 * time.Second
)

// Client représente un client pour l'API TCGdex
type Client struct {
	httpClient *http.Client
	baseURL    string
}

// NewClient crée une nouvelle instance du client TCGdex
func NewClient() *Client {
	return &Client{
		httpClient: &http.Client{
			Timeout: defaultTimeout,
		},
		baseURL: baseURL,
	}
}

// GetCard récupère les détails d'une carte
func (c *Client) GetCard(id string) (*models.Card, error) {
	url := fmt.Sprintf("%s/cards/%s", c.baseURL, id)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("erreur lors de la création de la requête: %w", err)
	}

	req.Header.Set("User-Agent", "Pokemon-TCG-Collection/1.0")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("erreur lors de la requête: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("status code inattendu: %d", resp.StatusCode)
	}

	var card models.Card
	if err := json.NewDecoder(resp.Body).Decode(&card); err != nil {
		return nil, fmt.Errorf("erreur lors du décodage de la réponse: %w", err)
	}

	// Set the image URL
	card.SetImage(c.baseURL)

	return &card, nil
}

// SearchCards recherche des cartes selon les critères donnés
func (c *Client) SearchCards(options models.SearchRequest) (*models.CardCollection, error) {
	url := fmt.Sprintf("%s/cards", c.baseURL)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("erreur lors de la création de la requête: %w", err)
	}

	// Ajout des paramètres de recherche
	q := req.URL.Query()
	if options.Query != "" {
		q.Add("q", options.Query)
	}
	if options.Series != "" {
		q.Add("series", options.Series)
	}
	q.Add("page", fmt.Sprintf("%d", options.Page))
	q.Add("pageSize", fmt.Sprintf("%d", options.PageSize))
	req.URL.RawQuery = q.Encode()

	req.Header.Set("User-Agent", "Pokemon-TCG-Collection/1.0")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("erreur lors de la requête: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("status code inattendu: %d", resp.StatusCode)
	}

	var cardList models.CardCollection
	if err := json.NewDecoder(resp.Body).Decode(&cardList); err != nil {
		return nil, fmt.Errorf("erreur lors du décodage de la réponse: %w", err)
	}

	// Set image URLs for all cards
	for i := range cardList.Cards {
		cardList.Cards[i].SetImage(c.baseURL)
	}

	return &cardList, nil
}

// GetSets récupère la liste des sets disponibles
func (c *Client) GetSets() ([]models.Set, error) {
	url := fmt.Sprintf("%s/sets", c.baseURL)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("erreur lors de la création de la requête: %w", err)
	}

	req.Header.Set("User-Agent", "Pokemon-TCG-Collection/1.0")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("erreur lors de la requête: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("status code inattendu: %d", resp.StatusCode)
	}

	var sets []models.Set
	if err := json.NewDecoder(resp.Body).Decode(&sets); err != nil {
		return nil, fmt.Errorf("erreur lors du décodage de la réponse: %w", err)
	}

	return sets, nil
}
