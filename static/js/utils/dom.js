// Gestion des images
const ImageUtils = {
    handleError: (img, placeholderSrc = '/static/img/card-placeholder.png') => {
        img.src = placeholderSrc;
    },

    handleLoad: (img) => {
        img.style.opacity = '1';
    },

    setupLazyLoading: (selector = '.card img') => {
        const images = document.querySelectorAll(selector);
        images.forEach(img => {
            img.style.opacity = '0';
            img.style.transition = 'opacity 0.3s ease-in-out';
            img.addEventListener('error', () => ImageUtils.handleError(img));
            img.addEventListener('load', () => ImageUtils.handleLoad(img));
        });
    }
};

// Gestion des modales
const ModalUtils = {
    create: (content) => {
        const modal = document.createElement('div');
        modal.className = 'modal';
        modal.innerHTML = content;
        document.body.appendChild(modal);
        setTimeout(() => modal.classList.add('active'), 10);
        return modal;
    },

    close: (button) => {
        const modal = button.closest('.modal');
        modal.classList.remove('active');
        setTimeout(() => modal.remove(), 300);
    },

    setupCloseEvents: (modal) => {
        modal.addEventListener('click', (e) => {
            if (e.target === modal) {
                ModalUtils.close(modal.querySelector('.close-modal'));
            }
        });

        document.addEventListener('keydown', (e) => {
            if (e.key === 'Escape') {
                const activeModal = document.querySelector('.modal.active');
                if (activeModal) {
                    ModalUtils.close(activeModal.querySelector('.close-modal'));
                }
            }
        });
    }
};

// Animations
const AnimationUtils = {
    fadeIn: (element, duration = 300) => {
        element.style.opacity = '0';
        element.style.transition = `opacity ${duration}ms ease-in-out`;
        setTimeout(() => element.style.opacity = '1', 10);
    },

    fadeOut: (element, duration = 300) => {
        element.style.opacity = '0';
        setTimeout(() => element.remove(), duration);
    }
};

export { ImageUtils, ModalUtils, AnimationUtils }; 