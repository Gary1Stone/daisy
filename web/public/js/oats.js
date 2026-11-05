/* oats.js */

// Automatically set aria-invalid
document.querySelectorAll('input, select, textarea').forEach(input => {
    let touched = false;
    const updateValidity = () => {
        if (!touched) return;
        if (input.validity.valid) {
            input.removeAttribute('aria-invalid');
        } else {
            input.setAttribute('aria-invalid', 'true');
        }
    };
    input.addEventListener('blur', () => {
        touched = true;
        updateValidity();
    });
    input.addEventListener('input', updateValidity);
    input.addEventListener('change', updateValidity);
});