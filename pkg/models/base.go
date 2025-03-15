package models

// PageData représente les données de pagination pour l'affichage
type PageData struct {
	Cards          []Card
	CurrentPage    int
	TotalPages     int
	HasNext        bool
	HasPrev        bool
	SelectedSeries string
}

// PokemonSeries représente une série de cartes Pokémon
type PokemonSeries struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	CardCount int    `json:"cardCount"`
}

// GetAllSeries retourne toutes les séries disponibles
func GetAllSeries() []PokemonSeries {
	return []PokemonSeries{
		{ID: "swsh1", Name: "Épée et Bouclier", CardCount: 202},
		{ID: "swsh2", Name: "Clash des Rebelles", CardCount: 192},
		{ID: "swsh3", Name: "Ténèbres Embrasées", CardCount: 189},
		{ID: "swsh4", Name: "Voltage Éclatant", CardCount: 185},
		{ID: "swsh5", Name: "Styles de Combat", CardCount: 163},
		{ID: "swsh6", Name: "Règne de Glace", CardCount: 198},
		{ID: "swsh7", Name: "Évolution Céleste", CardCount: 203},
		{ID: "swsh8", Name: "Poing de Fusion", CardCount: 264},
		{ID: "swsh9", Name: "Étoiles Brillantes", CardCount: 172},
		{ID: "swsh10", Name: "Origine Perdue", CardCount: 196},
	}
}

// GetSeriesByID retourne une série par son ID
func GetSeriesByID(id string) (PokemonSeries, bool) {
	for _, series := range GetAllSeries() {
		if series.ID == id {
			return series, true
		}
	}
	return PokemonSeries{}, false
}

// Card représente une carte Pokémon
type Card struct {
	ID       string   `json:"id"`
	Name     string   `json:"name"`
	Number   string   `json:"number"`
	Series   string   `json:"series"`
	Set      Set      `json:"set"`
	Types    []string `json:"types,omitempty"`
	HP       int      `json:"hp,omitempty"`
	Rarity   string   `json:"rarity,omitempty"`
	ImageURL string   `json:"image"`
	Image    string   `json:"-"` // URL de l'image pour l'affichage
}

// NewCard crée une nouvelle instance de Card
func NewCard(id, name string, set Set) Card {
	return Card{
		ID:   id,
		Name: name,
		Set:  set,
	}
}

// SetImage définit l'URL de l'image de la carte
func (c *Card) SetImage(baseURL string) {
	if c.ImageURL == "" {
		c.ImageURL = baseURL + "/cards/" + c.ID + "/image"
	}
}

// IsValid vérifie si la carte est valide
func (c *Card) IsValid() bool {
	return c.ID != "" && c.Name != ""
}

// Set représente un set de cartes
type Set struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Series string `json:"series"`
	Total  int    `json:"total"`
}
