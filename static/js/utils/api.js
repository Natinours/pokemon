// Configuration de l'API
const API_CONFIG = {
    baseUrl: '/api',
    headers: {
        'Content-Type': 'application/json'
    }
};

// Gestion des favoris
const FavoritesAPI = {
    add: async (cardId) => {
        try {
            const response = await fetch(`${API_CONFIG.baseUrl}/favorites/add`, {
                method: 'POST',
                headers: API_CONFIG.headers,
                body: JSON.stringify({ cardId })
            });
            return await response.json();
        } catch (error) {
            console.error('Erreur lors de l\'ajout aux favoris:', error);
            throw error;
        }
    },

    remove: async (cardId) => {
        try {
            const response = await fetch(`${API_CONFIG.baseUrl}/favorites/remove`, {
                method: 'POST',
                headers: API_CONFIG.headers,
                body: JSON.stringify({ cardId })
            });
            return await response.json();
        } catch (error) {
            console.error('Erreur lors de la suppression des favoris:', error);
            throw error;
        }
    },

    removeAll: async () => {
        try {
            const response = await fetch(`${API_CONFIG.baseUrl}/favorites/remove-all`, {
                method: 'POST',
                headers: API_CONFIG.headers
            });
            return await response.json();
        } catch (error) {
            console.error('Erreur lors de la suppression de tous les favoris:', error);
            throw error;
        }
    }
};

// Gestion des cartes
const CardsAPI = {
    search: async (query, options = {}) => {
        try {
            const params = new URLSearchParams({
                q: query,
                ...options
            });
            const response = await fetch(`${API_CONFIG.baseUrl}/cards/search?${params}`);
            return await response.json();
        } catch (error) {
            console.error('Erreur lors de la recherche:', error);
            throw error;
        }
    },

    getImage: (series, number, quality = 'high') => {
        return `/proxy/card-image/${series}/${number}/${quality}`;
    }
};

export { FavoritesAPI, CardsAPI }; 