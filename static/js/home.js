// Fonction de recherche
function search() {
    const searchInput = document.getElementById('searchInput');
    const query = searchInput.value.trim();
    
    if (query) {
        window.location.href = `/search?q=${encodeURIComponent(query)}`;
    }
}

// Gérer la soumission avec la touche Entrée
document.addEventListener('DOMContentLoaded', function() {
    const searchInput = document.getElementById('searchInput');
    searchInput.addEventListener('keypress', function(e) {
        if (e.key === 'Enter') {
            e.preventDefault();
            search();
        }
    });
}); 