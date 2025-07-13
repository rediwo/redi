// Utility functions for Svelte components

export function formatDate(date) {
    const d = new Date(date);
    return d.toLocaleDateString('en-US', {
        year: 'numeric',
        month: 'long',
        day: 'numeric'
    });
}

export function debounce(func, wait) {
    let timeout;
    return function executedFunction(...args) {
        const later = () => {
            clearTimeout(timeout);
            func(...args);
        };
        clearTimeout(timeout);
        timeout = setTimeout(later, wait);
    };
}

export const API_URL = '/api';

export const CONFIG = {
    theme: 'light',
    language: 'en',
    debug: true
};

// Helper to capitalize strings
export function capitalize(str) {
    return str.charAt(0).toUpperCase() + str.slice(1);
}

// Default export
export default {
    formatDate,
    debounce,
    capitalize,
    API_URL,
    CONFIG
};