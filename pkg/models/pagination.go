package models

// SearchRequest représente les paramètres de recherche de cartes
type SearchRequest struct {
	Query    string
	Series   string
	Page     int
	PageSize int
}

// PageInfo représente les informations de pagination pour l'affichage
type PageInfo struct {
	Cards          []Card
	CurrentPage    int
	TotalPages     int
	HasNext        bool
	HasPrev        bool
	SelectedSeries string
}

// CardCollection représente une liste paginée de cartes pour l'API
type CardCollection struct {
	Cards      []Card `json:"cards"`
	Page       int    `json:"page"`
	PageSize   int    `json:"pageSize"`
	TotalCards int    `json:"total"`
}
