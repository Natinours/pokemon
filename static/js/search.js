// Mapping des types pour la correspondance entre l'interface et l'API
const TYPE_MAPPING = {
    'Normal': 'Incolore',
    'Incolore': 'Normal'
};

// Effectuer la recherche
function performSearch() {
    const searchInput = document.getElementById('searchInput');
    if (!searchInput) return;
    
    const query = searchInput.value.trim();
    if (query) {
        window.location.href = `/search?q=${encodeURIComponent(query)}`;
    }
}

// Écouter la touche Entrée dans le champ de recherche
document.addEventListener('DOMContentLoaded', function() {
    const searchInput = document.getElementById('searchInput');
    if (searchInput) {
        searchInput.addEventListener('keypress', function(e) {
            if (e.key === 'Enter') {
                performSearch();
            }
        });
    }

    // Initialiser l'état des boutons favoris
    const favorites = getFavorites();
    document.querySelectorAll('.favorite-btn').forEach(button => {
        const cardId = button.getAttribute('data-card-id');
        if (favorites.some(card => card.id === cardId)) {
            button.classList.add('active');
            button.innerHTML = '♥ Retirer des favoris';
        }
    });
});

// Fonction pour appliquer les filtres
function applyFilters() {
    const typeFilter = document.getElementById('typeFilter');
    if (!typeFilter) return;

    const selectedType = typeFilter.value;
    console.log('Type sélectionné:', selectedType);
    
    let visibleCards = 0;
    document.querySelectorAll('.card').forEach(card => {
        // Récupérer les types depuis l'attribut data-types
        const cardTypesStr = card.getAttribute('data-types');
        console.log('Types bruts de la carte:', cardTypesStr);
        
        const cardTypes = cardTypesStr ? cardTypesStr.split(',').map(type => type.trim()) : [];
        console.log('Types de la carte après traitement:', cardTypes);
        
        // Vérifier si la carte a le type sélectionné
        const hasType = !selectedType || cardTypes.includes(selectedType);
        console.log('La carte correspond au filtre:', hasType, 'pour les types:', cardTypes, 'et le type sélectionné:', selectedType);
        
        card.style.display = hasType ? '' : 'none';
        if (hasType) visibleCards++;
    });
    
    console.log('Nombre de cartes visibles:', visibleCards);
}

// Gestion du cache des favoris
const FAVORITES_STORAGE_KEY = 'pokemon_favorites';

function getFavorites() {
    const favorites = localStorage.getItem(FAVORITES_STORAGE_KEY);
    return favorites ? JSON.parse(favorites) : [];
}

function addToFavorites(cardData) {
    const favorites = getFavorites();
    if (!favorites.some(card => card.id === cardData.id)) {
        favorites.push(cardData);
        localStorage.setItem(FAVORITES_STORAGE_KEY, JSON.stringify(favorites));
    }
}

function removeFromFavorites(cardId) {
    const favorites = getFavorites();
    const updatedFavorites = favorites.filter(card => card.id !== cardId);
    localStorage.setItem(FAVORITES_STORAGE_KEY, JSON.stringify(updatedFavorites));
}

// Gestion des favoris
function toggleFavorite(cardId, event) {
    event.preventDefault();
    event.stopPropagation();
    
    const card = event.target.closest('.card');
    const button = event.target;
    const isFavorite = button.classList.contains('active');

    // Récupérer les données de la carte
    const cardData = {
        id: cardId,
        name: card.querySelector('.card-title').textContent,
        image: card.querySelector('img').src,
        set: card.querySelector('.card-set').textContent
    };

    if (!isFavorite) {
        // Ajouter aux favoris
        addToFavorites(cardData);
        button.classList.add('active');
        button.innerHTML = '♥ Retirer des favoris';
        showNotification('Carte ajoutée aux favoris', true);
    } else {
        // Retirer des favoris
        removeFromFavorites(cardId);
        button.classList.remove('active');
        button.innerHTML = '♡ Ajouter aux favoris';
        showNotification('Carte retirée des favoris', false);
    }

    // Mettre à jour le statut sur le serveur
    fetch(`/api/favorites/${isFavorite ? 'remove' : 'add'}`, {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json'
        },
        body: JSON.stringify({ cardId })
    })
    .catch(error => {
        console.error('Erreur:', error);
        showNotification('Une erreur est survenue', false);
    });
}

// Afficher une notification
function showNotification(message, isSuccess) {
    // Supprimer la notification existante si elle existe
    const existingNotif = document.querySelector('.notification');
    if (existingNotif) {
        existingNotif.remove();
    }

    // Créer la nouvelle notification
    const notification = document.createElement('div');
    notification.className = `notification ${isSuccess ? 'success' : 'error'}`;
    notification.innerHTML = `
        ${message}
        ${isSuccess ? `
            <div class="notification-actions">
                <a href="/favorites" class="view-favorites-btn">Voir les favoris</a>
            </div>
        ` : ''}
    `;

    // Ajouter la notification au document
    document.body.appendChild(notification);

    // Animer l'apparition
    setTimeout(() => notification.classList.add('show'), 10);

    // Supprimer après 3 secondes
    setTimeout(() => {
        notification.classList.remove('show');
        setTimeout(() => notification.remove(), 300);
    }, 3000);
}

// Gestion de la modale
function openModal(series, number) {
    const modal = document.getElementById('imageModal');
    const modalImage = document.getElementById('modalImage');
    
    modalImage.src = `/proxy/card-image/${series}/${number}/high`;
    modal.classList.add('active');
}

function closeModal(event) {
    event.preventDefault();
    const modal = document.getElementById('imageModal');
    modal.classList.remove('active');
}

// Gestion des erreurs d'images
function handleImageError(img) {
    img.src = '/static/img/card-placeholder.png';
}

function handleImageLoad(img) {
    img.style.opacity = '1';
}

// Initialisation
document.addEventListener('DOMContentLoaded', function() {
    // Vérifier l'état initial des filtres
    applyFilters();

    // Gestion du chargement progressif des images
    const images = document.querySelectorAll('.card img');
    images.forEach(img => {
        img.style.opacity = '0';
        img.style.transition = 'opacity 0.3s ease-in-out';
    });

    // Fermer la modale avec la touche Echap
    document.addEventListener('keydown', (e) => {
        if (e.key === 'Escape') {
            const modal = document.getElementById('imageModal');
            if (modal.classList.contains('active')) {
                modal.classList.remove('active');
            }
        }
    });
}); 