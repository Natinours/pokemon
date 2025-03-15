// Gestion des favoris
function toggleFavorite(cardId) {
    const btn = document.querySelector(`[data-card-id="${cardId}"]`);
    const isActive = btn.classList.contains('active');
    
    fetch(`/api/favorites/${isActive ? 'remove' : 'add'}`, {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json'
        },
        body: JSON.stringify({ cardId: cardId })
    })
    .then(response => response.json())
    .then(data => {
        if (data.success) {
            btn.classList.toggle('active');
            btn.textContent = isActive ? '♡ Ajouter aux favoris' : '♥ Retirer des favoris';
        }
    })
    .catch(error => {
        console.error('Erreur:', error);
        alert('Une erreur est survenue lors de la mise à jour des favoris.');
    });
}

// Gestion de la pagination
function changePage(page) {
    const urlParams = new URLSearchParams(window.location.search);
    urlParams.set('page', page);
    window.location.search = urlParams.toString();
}

// Gestion de la recherche
function handleSearch(event) {
    event.preventDefault();
    const searchInput = document.querySelector('.search-input');
    const searchTerm = searchInput.value.trim();
    
    if (searchTerm) {
        window.location.href = `/search?q=${encodeURIComponent(searchTerm)}`;
    }
}

// Gestion du chargement des images
document.addEventListener('DOMContentLoaded', function() {
    const images = document.querySelectorAll('.card img');
    images.forEach(img => {
        img.addEventListener('error', function() {
            this.src = '/static/img/card-placeholder.png';
        });
    });
}); 