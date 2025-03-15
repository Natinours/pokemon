package handlers

import (
	"compress/gzip"
	"cours/pokemon/pkg/models"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

const (
	cardsPerPage  = 12
	apiBaseURL    = "https://api.tcgdex.net/v2/fr"
	cacheDuration = 24 * time.Hour         // Durée de validité du cache (24 heures)
	maxCacheSize  = 500                    // Nombre maximum d'entrées dans le cache
	maxCacheBytes = 100 * 1024 * 1024      // 100 MB maximum pour le cache d'images
	maxRetries    = 3                      // Nombre maximum de tentatives pour les requêtes API
	retryDelay    = 500 * time.Millisecond // Délai entre les tentatives
	maxWorkers    = 4                      // Nombre maximum de workers pour le chargement parallèle
	preloadPages  = 1                      // Nombre de pages à précharger
)

// Structure pour les données de la carte
type TCGCard struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Set  struct {
		Name string `json:"name"`
	} `json:"set"`
}

// Structure pour le cache des cartes
type CardCache struct {
	sync.RWMutex
	names       map[string]string    // Cache des noms de cartes
	images      map[string][]byte    // Cache des images
	lastUpdated map[string]time.Time // Horodatage de dernière mise à jour
	lastAccess  map[string]time.Time // Horodatage du dernier accès
	totalBytes  int64                // Taille totale des images en cache
}

// Métriques de performance
type Metrics struct {
	sync.RWMutex
	cacheHits     int64
	cacheMisses   int64
	apiErrors     int64
	avgApiLatency time.Duration
	apiCalls      int64
}

// Structure pour le préchargement
type PreloadQueue struct {
	sync.Mutex
	queue    map[string]struct{}
	inFlight map[string]struct{}
}

// Structure pour les requêtes API en batch
type BatchRequest struct {
	Series string
	Start  int
	End    int
}

// Structure pour la réponse du batch
type BatchResponse struct {
	Cards map[string]TCGCard
	Error error
}

var (
	cardCache = &CardCache{
		names:       make(map[string]string),
		images:      make(map[string][]byte),
		lastUpdated: make(map[string]time.Time),
		lastAccess:  make(map[string]time.Time),
		totalBytes:  0,
	}

	metrics = &Metrics{}

	preloadQueue = &PreloadQueue{
		queue:    make(map[string]struct{}),
		inFlight: make(map[string]struct{}),
	}

	// Canal pour les requêtes batch
	batchRequests = make(chan BatchRequest)
	batchResults  = make(chan BatchResponse)
)

// recordApiLatency enregistre la latence d'un appel API
func (m *Metrics) recordApiLatency(start time.Time) {
	m.Lock()
	defer m.Unlock()
	m.apiCalls++
	duration := time.Since(start)
	m.avgApiLatency = time.Duration((int64(m.avgApiLatency)*int64(m.apiCalls-1) + int64(duration)) / int64(m.apiCalls))
}

// recordCacheHit enregistre un succès du cache
func (m *Metrics) recordCacheHit() {
	atomic.AddInt64(&m.cacheHits, 1)
}

// recordCacheMiss enregistre un échec du cache
func (m *Metrics) recordCacheMiss() {
	atomic.AddInt64(&m.cacheMisses, 1)
}

// recordApiError enregistre une erreur API
func (m *Metrics) recordApiError() {
	atomic.AddInt64(&m.apiErrors, 1)
}

// fetchWithRetry effectue une requête HTTP avec retry
func fetchWithRetry(url string) (*http.Response, error) {
	var lastErr error
	for i := 0; i < maxRetries; i++ {
		if i > 0 {
			time.Sleep(retryDelay * time.Duration(i))
		}

		start := time.Now()
		resp, err := http.Get(url)
		if err == nil {
			metrics.recordApiLatency(start)
			return resp, nil
		}

		lastErr = err
		metrics.recordApiError()
		log.Printf("Tentative %d échouée pour %s: %v", i+1, url, err)
	}
	return nil, fmt.Errorf("échec après %d tentatives: %v", maxRetries, lastErr)
}

// Traitement des requêtes API en batch
func processBatchRequests() {
	for {
		// Attendre une requête
		request := <-batchRequests

		// Préparer la requête batch
		cards := make(map[string]TCGCard)

		// Construire l'URL pour le batch
		url := fmt.Sprintf("%s/cards/%s?page=1&pageSize=%d", apiBaseURL, request.Series, request.End-request.Start+1)

		// Faire la requête
		start := time.Now()
		resp, err := httpClient.Get(url)
		if err != nil {
			metrics.recordApiError()
			batchResults <- BatchResponse{nil, err}
			continue
		}

		// Lire et décoder la réponse
		var batchCards []TCGCard
		if err := json.NewDecoder(resp.Body).Decode(&batchCards); err != nil {
			resp.Body.Close()
			batchResults <- BatchResponse{nil, err}
			continue
		}
		resp.Body.Close()

		// Enregistrer la latence
		metrics.recordApiLatency(start)

		// Traiter les cartes reçues
		for i, card := range batchCards {
			cardNum := request.Start + i
			cardKey := fmt.Sprintf("%s-%d", request.Series, cardNum)
			cards[cardKey] = card
		}

		// Envoyer les résultats
		batchResults <- BatchResponse{cards, nil}
	}
}

// GetCardInfo avec support du batch
func (c *CardCache) GetCardInfo(series string, number int) models.Card {
	cardKey := fmt.Sprintf("%s-%d", series, number)
	numberStr := strconv.Itoa(number)
	seriesName := getSetName(series)

	// Vérifier dans le cache
	c.RLock()
	if name, exists := c.names[cardKey]; exists {
		if time.Since(c.lastUpdated[cardKey]) < cacheDuration {
			c.RUnlock()
			metrics.recordCacheHit()
			return models.Card{
				ID:   cardKey,
				Name: name,
				Set: models.Set{
					ID:   series,
					Name: seriesName,
				},
				Number: numberStr,
			}
		}
	}
	c.RUnlock()
	metrics.recordCacheMiss()

	// Construire l'URL en fonction de la série
	var urls []string
	paddedNumber := fmt.Sprintf("%03d", number)

	if series == "swsh9" || series == "swsh10" {
		// Formats spéciaux pour swsh9 et swsh10
		urls = append(urls,
			fmt.Sprintf("%s/cards/%s/%s", apiBaseURL, series, paddedNumber),
			fmt.Sprintf("%s/cards/%s/%d", apiBaseURL, series, number),
			fmt.Sprintf("%s/cards/%s-%s", apiBaseURL, series, paddedNumber),
			fmt.Sprintf("%s/cards/%s-%d", apiBaseURL, series, number),
			fmt.Sprintf("%s/cards/swsh/%s/%s", apiBaseURL, series[4:], paddedNumber),
			fmt.Sprintf("%s/cards/swsh/%s/%d", apiBaseURL, series[4:], number))
	} else {
		// Format standard pour les autres séries
		urls = append(urls,
			fmt.Sprintf("%s/cards/%s-%s", apiBaseURL, series, numberStr),
			fmt.Sprintf("%s/cards/%s/%s", apiBaseURL, series, numberStr))
	}

	// Essayer chaque URL
	var card TCGCard
	var cardFound bool
	for _, url := range urls {
		log.Printf("Tentative de récupération de la carte depuis %s", url)
		resp, err := httpClient.Get(url)
		if err != nil {
			log.Printf("Erreur lors de la requête à %s: %v", url, err)
			continue
		}

		if resp.StatusCode != http.StatusOK {
			resp.Body.Close()
			log.Printf("Status code non-OK (%d) pour %s", resp.StatusCode, url)
			continue
		}

		var tempCard TCGCard
		if err := json.NewDecoder(resp.Body).Decode(&tempCard); err != nil {
			resp.Body.Close()
			log.Printf("Erreur de décodage pour %s: %v", url, err)
			continue
		}
		resp.Body.Close()

		// Si on a trouvé un nom valide
		if tempCard.Name != "" {
			card = tempCard
			cardFound = true
			break
		}
	}

	// Si on n'a pas trouvé de nom valide, essayer de récupérer la liste complète des cartes
	if !cardFound || card.Name == "" {
		// Faire une requête à l'API pour obtenir la liste des cartes de la série
		seriesURL := fmt.Sprintf("%s/series/%s/cards", apiBaseURL, series)
		resp, err := httpClient.Get(seriesURL)
		if err == nil && resp.StatusCode == http.StatusOK {
			var cards []TCGCard
			if err := json.NewDecoder(resp.Body).Decode(&cards); err == nil {
				// Chercher la carte avec le bon numéro
				for _, c := range cards {
					if strings.HasSuffix(c.ID, fmt.Sprintf("-%d", number)) ||
						strings.HasSuffix(c.ID, fmt.Sprintf("-%s", paddedNumber)) {
						card = c
						cardFound = true
						break
					}
				}
			}
			resp.Body.Close()
		}
	}

	// Si toujours pas de nom, utiliser un nom par défaut
	if !cardFound || card.Name == "" {
		card.Name = fmt.Sprintf("Carte %s #%s", seriesName, numberStr)
	}

	// Mettre en cache
	c.Lock()
	c.names[cardKey] = card.Name
	c.lastUpdated[cardKey] = time.Now()
	c.lastAccess[cardKey] = time.Now()
	c.Unlock()

	return models.Card{
		ID:   cardKey,
		Name: card.Name,
		Set: models.Set{
			ID:   series,
			Name: seriesName,
		},
		Number: numberStr,
	}
}

// cleanCache nettoie le cache si nécessaire
func (c *CardCache) cleanCache() {
	c.Lock()
	defer c.Unlock()

	// Si le cache n'a pas atteint ses limites, ne rien faire
	if len(c.images) < maxCacheSize && c.totalBytes < int64(maxCacheBytes) {
		return
	}

	// Créer une liste des entrées triées par date de dernier accès
	type cacheEntry struct {
		key        string
		lastAccess time.Time
		size       int64
	}
	entries := make([]cacheEntry, 0, len(c.lastAccess))
	for key, access := range c.lastAccess {
		if img, exists := c.images[key]; exists {
			entries = append(entries, cacheEntry{key, access, int64(len(img))})
		}
	}

	// Trier les entrées par date de dernier accès (les plus anciennes en premier)
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].lastAccess.Before(entries[j].lastAccess)
	})

	// Supprimer les entrées jusqu'à ce que nous soyons sous les limites
	targetSize := int64(maxCacheBytes) * 8 / 10 // Viser 80% de la limite
	for i := 0; i < len(entries) && (len(c.images) > maxCacheSize*8/10 || c.totalBytes > targetSize); i++ {
		key := entries[i].key
		c.totalBytes -= entries[i].size
		delete(c.images, key)
		delete(c.lastAccess, key)
		delete(c.lastUpdated, key)
	}
}

// updateLastAccess met à jour l'horodatage du dernier accès de manière thread-safe
func (c *CardCache) updateLastAccess(key string) {
	c.Lock()
	if _, exists := c.images[key]; exists {
		c.lastAccess[key] = time.Now()
	}
	c.Unlock()
}

// GetCardImage récupère une image depuis le cache ou l'API
func (c *CardCache) GetCardImage(series, number, quality string) ([]byte, error) {
	imageKey := fmt.Sprintf("%s-%s-%s", series, number, quality)

	// Vérifier dans le cache
	c.RLock()
	if img, exists := c.images[imageKey]; exists {
		if time.Since(c.lastUpdated[imageKey]) < cacheDuration {
			c.RUnlock()
			c.updateLastAccess(imageKey)
			return img, nil
		}
	}
	c.RUnlock()

	// Nettoyer le cache si nécessaire
	c.cleanCache()

	// Format du numéro selon la série
	var paddedNumber string
	if series == "swsh10" {
		// Pour swsh10, on veut un format 060
		if len(number) == 1 {
			paddedNumber = "00" + number
		} else if len(number) == 2 {
			paddedNumber = "0" + number
		} else {
			paddedNumber = number
		}
	} else {
		// Pour les autres séries, on garde le numéro tel quel
		paddedNumber = number
	}

	// Nouveau format d'URL basé sur la structure correcte de l'API
	urls := []string{
		fmt.Sprintf("https://assets.tcgdex.net/fr/swsh/%s/%s/%s.webp", series, paddedNumber, quality),
		fmt.Sprintf("https://assets.tcgdex.net/fr/swsh/%s/%s/%s.png", series, paddedNumber, quality),
	}

	// Essayer toutes les URLs possibles
	var lastErr error
	for _, url := range urls {
		img, err := fetchImage(url)
		if err == nil {
			// Mettre en cache avec suivi de la taille
			c.Lock()
			// Soustraire la taille de l'ancienne image si elle existe
			if oldImg, exists := c.images[imageKey]; exists {
				c.totalBytes -= int64(len(oldImg))
			}
			c.images[imageKey] = img
			c.totalBytes += int64(len(img))
			c.lastUpdated[imageKey] = time.Now()
			c.lastAccess[imageKey] = time.Now()
			c.Unlock()
			return img, nil
		}
		lastErr = err
		log.Printf("Échec de récupération de l'image depuis %s: %v", url, err)
	}

	// Si la qualité demandée est "high", essayer automatiquement en "low"
	if quality == "high" {
		lowQualityImg, err := c.GetCardImage(series, number, "low")
		if err == nil {
			return lowQualityImg, nil
		}
	}

	return nil, fmt.Errorf("aucune image trouvée pour %s/%s (dernière erreur: %v)", series, number, lastErr)
}

// Fonction utilitaire pour récupérer une image
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

// Fonction utilitaire pour obtenir le nom du set
func getSetName(series string) string {
	if seriesInfo, ok := models.GetSeriesByID(series); ok {
		return seriesInfo.Name
	}
	return series
}

// Structure pour les résultats du chargement parallèle
type cardResult struct {
	card  models.Card
	index int
}

// Client HTTP réutilisable avec timeout et keep-alive
var httpClient = &http.Client{
	Timeout: 10 * time.Second,
	Transport: &http.Transport{
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: maxWorkers,
		IdleConnTimeout:     90 * time.Second,
	},
}

// Ajouter une image à la file de préchargement
func (pq *PreloadQueue) Add(series string, number int) {
	key := fmt.Sprintf("%s-%d", series, number)
	pq.Lock()
	defer pq.Unlock()

	if _, exists := pq.inFlight[key]; exists {
		return
	}
	pq.queue[key] = struct{}{}
}

// Démarrer le préchargement d'une image
func (pq *PreloadQueue) StartPreload(series string, number int) bool {
	key := fmt.Sprintf("%s-%d", series, number)
	pq.Lock()
	defer pq.Unlock()

	if _, exists := pq.inFlight[key]; exists {
		return false
	}
	delete(pq.queue, key)
	pq.inFlight[key] = struct{}{}
	return true
}

// Terminer le préchargement d'une image
func (pq *PreloadQueue) FinishPreload(series string, number int) {
	key := fmt.Sprintf("%s-%d", series, number)
	pq.Lock()
	defer pq.Unlock()
	delete(pq.inFlight, key)
}

// Précharger les images en arrière-plan
func preloadImages() {
	for {
		time.Sleep(100 * time.Millisecond)

		preloadQueue.Lock()
		if len(preloadQueue.queue) == 0 {
			preloadQueue.Unlock()
			continue
		}

		// Sélectionner une image à précharger
		var series string
		var number int
		for key := range preloadQueue.queue {
			parts := strings.Split(key, "-")
			if len(parts) != 2 {
				delete(preloadQueue.queue, key)
				continue
			}
			series = parts[0]
			number, _ = strconv.Atoi(parts[1])
			break
		}
		preloadQueue.Unlock()

		if !preloadQueue.StartPreload(series, number) {
			continue
		}

		// Précharger l'image en basse qualité
		go func(s string, n int) {
			defer preloadQueue.FinishPreload(s, n)

			numberStr := strconv.Itoa(n)
			_, err := cardCache.GetCardImage(s, numberStr, "low")
			if err != nil {
				log.Printf("Erreur lors du préchargement de l'image %s/%d: %v", s, n, err)
			}
		}(series, number)
	}
}

// loadCardsParallel charge les cartes en parallèle avec préchargement
func loadCardsParallel(series string, startCard, endCard int) []models.Card {
	numCards := endCard - startCard + 1
	cards := make([]models.Card, numCards)

	// Créer un canal pour les résultats
	results := make(chan cardResult, numCards)

	// Créer un pool de workers
	workers := make(chan struct{}, maxWorkers)

	// Lancer les goroutines pour charger les cartes
	var wg sync.WaitGroup
	for i := 0; i < numCards; i++ {
		wg.Add(1)
		go func(cardNum, index int) {
			defer wg.Done()

			// Acquérir un worker
			workers <- struct{}{}
			defer func() { <-workers }()

			// Charger la carte
			card := cardCache.GetCardInfo(series, cardNum)
			card.Image = fmt.Sprintf("/proxy/card-image/%s/%d", series, cardNum)

			// Précharger les images suivantes
			if cardNum < endCard {
				preloadQueue.Add(series, cardNum+1)
			}

			// Envoyer le résultat
			results <- cardResult{card: card, index: index}
		}(startCard+i, i)
	}

	// Goroutine pour fermer le canal des résultats
	go func() {
		wg.Wait()
		close(results)
	}()

	// Collecter les résultats
	for result := range results {
		cards[result.index] = result.card
	}

	return cards
}

// HandleCollection avec préchargement des pages suivantes
func HandleCollection(w http.ResponseWriter, r *http.Request) {
	selectedSeries := r.URL.Query().Get("series")
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}

	log.Printf("Série sélectionnée : %s, Page : %d", selectedSeries, page)

	// Calculer le nombre total de cartes et de pages
	var totalCards int
	if selectedSeries == "" {
		// Somme des cartes de toutes les séries
		for _, seriesInfo := range models.GetAllSeries() {
			totalCards += seriesInfo.CardCount
		}
	} else {
		// Nombre de cartes de la série sélectionnée
		if seriesInfo, ok := models.GetSeriesByID(selectedSeries); ok {
			totalCards = seriesInfo.CardCount
		}
	}

	// Calculer la pagination
	totalPages := (totalCards + cardsPerPage - 1) / cardsPerPage
	if page > totalPages {
		page = totalPages
	}

	startIndex := (page - 1) * cardsPerPage
	endIndex := startIndex + cardsPerPage
	if endIndex > totalCards {
		endIndex = totalCards
	}

	var cards []models.Card
	if selectedSeries == "" {
		// Charger les cartes de toutes les séries pour la page courante
		currentCount := 0
		for _, seriesInfo := range models.GetAllSeries() {
			if currentCount+seriesInfo.CardCount <= startIndex {
				currentCount += seriesInfo.CardCount
				continue
			}

			seriesStartCard := 1
			if currentCount < startIndex {
				seriesStartCard = startIndex - currentCount + 1
			}

			seriesEndCard := seriesInfo.CardCount
			if currentCount+seriesInfo.CardCount > endIndex {
				seriesEndCard = endIndex - currentCount
			}

			// Charger les cartes de cette série en parallèle
			seriesCards := loadCardsParallel(seriesInfo.ID, seriesStartCard, seriesEndCard)
			cards = append(cards, seriesCards...)

			currentCount += seriesInfo.CardCount
			if currentCount >= endIndex {
				break
			}
		}
	} else {
		// Charger uniquement les cartes de la série sélectionnée en parallèle
		startCard := startIndex + 1
		endCard := endIndex
		if endCard > totalCards {
			endCard = totalCards
		}
		cards = loadCardsParallel(selectedSeries, startCard, endCard)
	}

	log.Printf("Cartes chargées pour la page %d : %d cartes", page, len(cards))

	// Précharger la page suivante si elle existe
	if page < totalPages {
		nextStartIndex := endIndex
		nextEndIndex := nextStartIndex + cardsPerPage
		if nextEndIndex > totalCards {
			nextEndIndex = totalCards
		}

		if selectedSeries == "" {
			// Précharger les cartes de la page suivante pour toutes les séries
			currentCount := 0
			for _, seriesInfo := range models.GetAllSeries() {
				if currentCount+seriesInfo.CardCount <= nextStartIndex {
					currentCount += seriesInfo.CardCount
					continue
				}

				seriesStartCard := 1
				if currentCount < nextStartIndex {
					seriesStartCard = nextStartIndex - currentCount + 1
				}

				seriesEndCard := seriesInfo.CardCount
				if currentCount+seriesInfo.CardCount > nextEndIndex {
					seriesEndCard = nextEndIndex - currentCount
				}

				// Ajouter les cartes à la file de préchargement
				for cardNum := seriesStartCard; cardNum <= seriesEndCard; cardNum++ {
					preloadQueue.Add(seriesInfo.ID, cardNum)
				}

				currentCount += seriesInfo.CardCount
				if currentCount >= nextEndIndex {
					break
				}
			}
		} else {
			// Précharger les cartes de la page suivante pour la série sélectionnée
			nextStartCard := endIndex + 1
			nextEndCard := nextEndIndex
			for cardNum := nextStartCard; cardNum <= nextEndCard; cardNum++ {
				preloadQueue.Add(selectedSeries, cardNum)
			}
		}
	}

	// Création des fonctions pour le template
	funcMap := template.FuncMap{
		"sub": func(a, b int) int {
			return a - b
		},
		"add": func(a, b int) int {
			return a + b
		},
		"split":      strings.Split,
		"getSetName": getSetName,
	}

	// Parsing du template avec les fonctions
	tmpl, err := template.New("layout.html").Funcs(funcMap).ParseFiles(
		filepath.Join("templates", "layout.html"),
		filepath.Join("templates", "collection.html"),
	)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	data := models.PageData{
		Cards:          cards,
		CurrentPage:    page,
		TotalPages:     totalPages,
		HasNext:        page < totalPages,
		HasPrev:        page > 1,
		SelectedSeries: selectedSeries,
	}

	// Activer la compression gzip si le navigateur la supporte
	if strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
		w.Header().Set("Content-Encoding", "gzip")
		w.Header().Set("Vary", "Accept-Encoding")
		gz := gzip.NewWriter(w)
		defer gz.Close()
		tmpl.ExecuteTemplate(gz, "layout", data)
		return
	}

	tmpl.ExecuteTemplate(w, "layout", data)
}

// HandleCardDetails gère l'affichage des détails d'une carte
func HandleCardDetails(w http.ResponseWriter, r *http.Request) {
	cardID := strings.TrimPrefix(r.URL.Path, "/card/")

	resp, err := http.Get(fmt.Sprintf("%s/cards/%s", apiBaseURL, cardID))
	if err != nil {
		http.Error(w, "Carte non trouvée", http.StatusNotFound)
		return
	}
	defer resp.Body.Close()

	var card models.Card
	if err := json.NewDecoder(resp.Body).Decode(&card); err != nil {
		http.Error(w, "Erreur lors du décodage de la carte", http.StatusInternalServerError)
		return
	}

	tmpl, err := template.ParseFiles(
		filepath.Join("templates", "layout.html"),
		filepath.Join("templates", "card.html"),
	)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	tmpl.ExecuteTemplate(w, "layout", card)
}

// HandleSearch gère la recherche de cartes
func HandleSearch(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	if query == "" {
		tmpl, err := template.New("layout.html").Funcs(template.FuncMap{
			"split": strings.Split,
		}).ParseFiles(
			filepath.Join("templates", "layout.html"),
			filepath.Join("templates", "search.html"),
		)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		tmpl.ExecuteTemplate(w, "layout", nil)
		return
	}

	var allResults []models.Card

	// Normaliser la requête de recherche
	normalizedQuery := strings.ToLower(query)
	for accent, normal := range map[string]string{
		"é": "e", "è": "e", "ê": "e", "ë": "e",
		"à": "a", "â": "a", "ä": "a",
		"î": "i", "ï": "i",
		"ô": "o", "ö": "o",
		"ù": "u", "û": "u", "ü": "u",
		"ÿ": "y",
		"ç": "c",
	} {
		normalizedQuery = strings.ReplaceAll(normalizedQuery, accent, normal)
	}

	// Faire une requête pour chaque set SWSH1 à SWSH10
	for i := 1; i <= 10; i++ {
		setID := fmt.Sprintf("swsh%d", i)
		searchURL := fmt.Sprintf("%s/sets/%s", apiBaseURL, setID)
		log.Printf("Tentative de récupération des cartes du set %s depuis: %s", setID, searchURL)

		resp, err := http.Get(searchURL)
		if err != nil {
			log.Printf("Erreur lors de la recherche pour le set %s: %v", setID, err)
			continue
		}

		// Vérifier le code de statut
		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			resp.Body.Close()
			log.Printf("Statut non-OK (%d) pour le set %s: %s", resp.StatusCode, setID, string(body))
			continue
		}

		// Décoder la réponse JSON
		var setInfo struct {
			ID    string `json:"id"`
			Name  string `json:"name"`
			Cards []struct {
				ID     string   `json:"id"`
				Name   string   `json:"name"`
				Number string   `json:"localId"`
				Types  []string `json:"types"`
			} `json:"cards"`
		}

		if err := json.NewDecoder(resp.Body).Decode(&setInfo); err != nil {
			resp.Body.Close()
			log.Printf("Erreur lors du décodage pour le set %s: %v", setID, err)
			continue
		}
		resp.Body.Close()

		// Parcourir les cartes du set
		for _, card := range setInfo.Cards {
			normalizedName := strings.ToLower(card.Name)
			for accent, normal := range map[string]string{
				"é": "e", "è": "e", "ê": "e", "ë": "e",
				"à": "a", "â": "a", "ä": "a",
				"î": "i", "ï": "i",
				"ô": "o", "ö": "o",
				"ù": "u", "û": "u", "ü": "u",
				"ÿ": "y",
				"ç": "c",
			} {
				normalizedName = strings.ReplaceAll(normalizedName, accent, normal)
			}

			if strings.Contains(normalizedName, normalizedQuery) {
				allResults = append(allResults, models.Card{
					ID:     card.ID,
					Name:   card.Name,
					Number: card.Number,
					Types:  card.Types,
					Image:  fmt.Sprintf("/proxy/card-image/%s/%s/low", setID, card.Number),
					Set: models.Set{
						ID:   setID,
						Name: setInfo.Name,
					},
				})
			}
		}
	}

	// Trier les résultats par nom
	sort.Slice(allResults, func(i, j int) bool {
		return allResults[i].Name < allResults[j].Name
	})

	data := struct {
		Query   string
		Results []models.Card
	}{
		Query:   query,
		Results: allResults,
	}

	tmpl, err := template.New("layout.html").Funcs(template.FuncMap{
		"split": strings.Split,
	}).ParseFiles(
		filepath.Join("templates", "layout.html"),
		filepath.Join("templates", "search.html"),
	)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Activer la compression gzip si le navigateur la supporte
	if strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
		w.Header().Set("Content-Encoding", "gzip")
		w.Header().Set("Vary", "Accept-Encoding")
		gz := gzip.NewWriter(w)
		defer gz.Close()
		tmpl.ExecuteTemplate(gz, "layout", data)
		return
	}

	tmpl.ExecuteTemplate(w, "layout", data)
}

// HandleCardImageProxy sert de proxy pour les images des cartes
func HandleCardImageProxy(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 4 {
		http.Error(w, "URL invalide", http.StatusBadRequest)
		return
	}

	// Extraire les parties de l'URL
	series := parts[len(parts)-3]  // swsh1 ou swsh2
	number := parts[len(parts)-2]  // numéro de la carte
	quality := parts[len(parts)-1] // "high" ou "low"

	// Si la qualité n'est pas spécifiée ou invalide, utiliser "low"
	if quality != "high" && quality != "low" {
		number = parts[len(parts)-1]
		series = parts[len(parts)-2]
		quality = "low"
	}

	// Récupérer l'image depuis le cache
	img, err := cardCache.GetCardImage(series, number, quality)
	if err != nil {
		http.Error(w, "Image non trouvée", http.StatusNotFound)
		return
	}

	// Détecter le type de contenu
	contentType := http.DetectContentType(img)
	if strings.HasPrefix(contentType, "image/") {
		w.Header().Set("Content-Type", contentType)
	} else {
		// Par défaut, utiliser PNG
		w.Header().Set("Content-Type", "image/png")
	}

	// Définir les en-têtes de cache pour le navigateur
	w.Header().Set("Cache-Control", "public, max-age=86400") // Cache pendant 24 heures
	w.Write(img)
}

// GetMetrics retourne les métriques actuelles
func GetMetrics() map[string]interface{} {
	metrics.RLock()
	defer metrics.RUnlock()

	cacheHitRate := float64(0)
	totalRequests := metrics.cacheHits + metrics.cacheMisses
	if totalRequests > 0 {
		cacheHitRate = float64(metrics.cacheHits) / float64(totalRequests) * 100
	}

	return map[string]interface{}{
		"cache_hit_rate":  fmt.Sprintf("%.2f%%", cacheHitRate),
		"cache_hits":      metrics.cacheHits,
		"cache_misses":    metrics.cacheMisses,
		"api_errors":      metrics.apiErrors,
		"avg_api_latency": metrics.avgApiLatency.String(),
		"total_api_calls": metrics.apiCalls,
	}
}

func init() {
	// Démarrer la goroutine de préchargement
	go preloadImages()
	// Démarrer les goroutines de traitement
	go processBatchRequests()
}
