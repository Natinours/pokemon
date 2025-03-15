// Animation des cartes au survol
document.addEventListener('DOMContentLoaded', function() {
    const cards = document.querySelectorAll('.card');
    
    cards.forEach(card => {
        card.addEventListener('mouseenter', function() {
            this.querySelector('.card-image-container img').style.transform = 'scale(1.1)';
        });

        card.addEventListener('mouseleave', function() {
            this.querySelector('.card-image-container img').style.transform = 'scale(1)';
        });
    });
});

// Gestion du chargement des images
document.addEventListener('DOMContentLoaded', function() {
    const images = document.querySelectorAll('.series-image');
    images.forEach(img => {
        img.addEventListener('error', function() {
            this.src = '/static/img/series-placeholder.png';
        });
    });
}); 