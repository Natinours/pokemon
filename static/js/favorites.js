// Supprimer tous les favoris
function removeAllFavorites() {
    if (!confirm('Êtes-vous sûr de vouloir supprimer tous vos favoris ?')) {
        return;
    }

    fetch('/api/favorites/remove-all', {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json'
        }
    })
    .then(response => response.json())
    .then(data => {
        if (data.success) {
            window.location.reload();
        } else {
            alert('Une erreur est survenue lors de la suppression des favoris.');
        }
    })
    .catch(error => {
        console.error('Erreur:', error);
        alert('Une erreur est survenue lors de la suppression des favoris.');
    });
}

// Exporter la liste des favoris
function exportFavorites() {
    const cards = Array.from(document.querySelectorAll('.card')).map(card => {
        const title = card.querySelector('.card-title').textContent;
        const set = card.querySelector('.card-info').textContent;
        const types = Array.from(card.querySelectorAll('.type-badge')).map(badge => badge.textContent).join(', ');
        return `${title} (${set}) - Types: ${types}`;
    });

    const content = cards.join('\n');
    const blob = new Blob([content], { type: 'text/plain' });
    const url = window.URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = 'favoris_pokemon.txt';
    document.body.appendChild(a);
    a.click();
    window.URL.revokeObjectURL(url);
    document.body.removeChild(a);
}

// Animation de suppression d'une carte
function animateRemoval(cardId) {
    const card = document.querySelector(`[data-card-id="${cardId}"]`);
    card.classList.add('removing');
    setTimeout(() => {
        card.remove();
        if (document.querySelectorAll('.card').length === 0) {
            location.reload();
        }
    }, 300);
}

// Gestion des favoris
function toggleFavorite(cardId) {
    const button = document.querySelector(`button[data-card-id="${cardId}"]`);
    const isFavorite = button.classList.contains('is-favorite');

    fetch(`/api/favorites/${isFavorite ? 'remove' : 'add'}`, {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json'
        },
        body: JSON.stringify({ cardId })
    })
    .then(response => response.json())
    .then(data => {
        if (data.status === 'success') {
            animateRemoval(cardId);
        }
    })
    .catch(error => console.error('Erreur:', error));
} 