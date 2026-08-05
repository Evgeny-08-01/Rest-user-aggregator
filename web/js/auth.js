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
    const userNameDisplay = document.getElementById('userNameDisplay');
    const userRoleDisplay = document.getElementById('userRoleDisplay');

    if (token && user) {
        loginForm.style.display = 'none';
        mainApp.style.display = 'block';
        userNameDisplay.textContent = user.username || 'Пользователь';
        userRoleDisplay.textContent = user.role || 'user';
        userRoleDisplay.className = 'badge ' + (user.role === 'admin' ? 'badge-admin' : 'badge-user');
    } else {
        loginForm.style.display = 'block';
        mainApp.style.display = 'none';
    }
}

// ============================================================
// ЛОГИН (используем функцию из api.js)
// ============================================================
// !!! ВАЖНО: УБИРАЕМ ДУБЛИРУЮЩУЮ ФУНКЦИЮ login !!!
// Раньше здесь была своя реализация login, которая дублировала
// функцию из api.js. Теперь мы просто экспортируем функцию из api.js
// под тем же именем, чтобы не ломать код в других файлах.
//
// Это называется "реэкспорт" (re-export) — мы берём функцию из api.js
// и делаем её доступной через auth.js.
// ============================================================
export function login(username, password) {
    // Просто вызываем функцию из api.js и возвращаем результат
    return apiLogin(username, password);
}