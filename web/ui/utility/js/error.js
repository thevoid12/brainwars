document.addEventListener('DOMContentLoaded', () => {
    const errorPopup = document.getElementById('errorPopup');
    const errorMessage = document.getElementById('errorMessage');
    const closeError = document.getElementById('closeError');

    function showError(message) {
        errorMessage.textContent = message;
        errorPopup.classList.add('show');
        setTimeout(hideError, 10000);
    }

    function hideError() {
        errorPopup.classList.remove('show');
    }

    closeError.addEventListener('click', hideError);

    const serverError = errorPopup?.dataset?.error?.trim();
    if (serverError) {
        showError(serverError);
    }
});
