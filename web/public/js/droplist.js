document.addEventListener("DOMContentLoaded", () => {
    // Scope all logic within a loop for multiple drop-list instances
    document.querySelectorAll('details[role="list"]').forEach(dropdown => {
        const container = dropdown.closest('.custom-select-container') || dropdown.parentElement;
        const hiddenInput = container.querySelector('input[type="hidden"]');
        const summary = dropdown.querySelector('summary');
        const options = Array.from(dropdown.querySelectorAll('ul[role="listbox"] a'));
        let activeIndex = -1;

        // 1. Initialize Default Value for this specific instance
        function initDefault() {
            const defaultVal = hiddenInput.value;
            if (defaultVal) {
                const match = options.find(opt => opt.getAttribute('data-value') === defaultVal);
                if (match) selectOption(match);
            }
        }

        // 2. Selection Logic
        function selectOption(optionEl) {
            options.forEach(opt => opt.removeAttribute('aria-selected'));

            optionEl.setAttribute('aria-selected', 'true');
            // Use innerHTML instead of textContent to preserve SVG icons
            summary.innerHTML = optionEl.innerHTML;
            hiddenInput.value = optionEl.getAttribute('data-value');
            dropdown.removeAttribute('open');
            hiddenInput.dispatchEvent(new Event('change', { bubbles: true })); // Dispatch change event
            summary.focus();
        }

        // 3. Click Listeners
        options.forEach(option => {
            option.addEventListener('click', (e) => {
                e.preventDefault();
                selectOption(option);
            });
        });

        // 4. Keyboard Navigation (scoped strictly to this dropdown)
        dropdown.addEventListener('keydown', (e) => {
            const isOpen = dropdown.hasAttribute('open');

            if (!isOpen) {
                if (e.key === 'ArrowDown' || e.key === 'Enter' || e.key === ' ') {
                    e.preventDefault();
                    dropdown.setAttribute('open', '');
                    activeIndex = options.findIndex(opt => opt.getAttribute('aria-selected') === 'true');
                    if (activeIndex === -1) activeIndex = 0;
                    options[activeIndex].focus();
                }
                return;
            }

            if (e.key === 'Escape') {
                e.preventDefault();
                dropdown.removeAttribute('open');
                summary.focus();
            } else if (e.key === 'ArrowDown') {
                e.preventDefault();
                activeIndex = (activeIndex + 1) % options.length;
                options[activeIndex].focus();
            } else if (e.key === 'ArrowUp') {
                e.preventDefault();
                activeIndex = (activeIndex - 1 + options.length) % options.length;
                options[activeIndex].focus();
            } else if (e.key === 'Enter' || e.key === ' ') {
                e.preventDefault();
                if (document.activeElement.tagName === 'A') {
                    selectOption(document.activeElement);
                }
            }
        });

        // Run initializer for this specific component
        initDefault();
    });
});