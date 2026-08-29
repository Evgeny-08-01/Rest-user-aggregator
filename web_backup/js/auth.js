// ============================================================
// АВТОРИЗАЦИЯ (JWT + РОЛИ)
// ============================================================

import { showToast } from './utils.js';
// !!! ВАЖНО: Импортируем apiFetch И login из api.js !!!
import { apiFetch, login as apiLogin } from './api.js';



const TOKEN_KEY = 'jwt_token';
const USER_KEY = 'jwt_user';

export function getToken() {
    return localStorage.getItem(TOKEN_KEY);
}

export function getUser() {
    try {
        return JSON.parse(localStorage.getItem(USER_KEY));
    } catch {
        return null;
    }
}

export function isAuthenticated() {
    return !!getToken();
}

export function isAdmin() {
    const user = getUser();
    return user && user.role === 'admin';
}

export function logout() {
    localStorage.removeItem(TOKEN_KEY);
    localStorage.removeItem(USER_KEY);
    window.location.reload();
}
export function checkAuthAndRender() {
    const token = getToken();
    const user = getUser();

    const loginForm = document.getElementById('loginForm');
    const mainApp = document.getElementById('mainApp');
    const registerForm = document.getElementById('registerForm');
    const userNameDisplay = document.getElementById('userNameDisplay');
    const userRoleDisplay = document.getElementById('userRoleDisplay');

    console.log('userNameDisplay:', userNameDisplay);
    console.log('user:', user);

    if (token && user) {
        if (loginForm) loginForm.style.display = 'none';
        if (mainApp) mainApp.style.display = 'block';
        if (registerForm) registerForm.style.display = 'none';
        if (userNameDisplay) userNameDisplay.textContent = user.email || 'Пользователь';
        if (userRoleDisplay) {
            userRoleDisplay.textContent = user.role || 'user';
            userRoleDisplay.className = 'badge ' + (user.role === 'admin' ? 'badge-admin' : 'badge-user');
        }
    } else {
        if (loginForm) loginForm.style.display = 'block';
        if (mainApp) mainApp.style.display = 'none';
        if (registerForm) registerForm.style.display = 'block';
        if (userNameDisplay) userNameDisplay.textContent = 'Гость';
        if (userRoleDisplay) {
            userRoleDisplay.textContent = 'user';
            userRoleDisplay.className = 'badge';
        }
    }
}