// ============================================================
// КОМПОНЕНТЫ (пагинация на фронтенде)
// ============================================================

import { escapeHtml, showToast } from './utils.js';
import { apiFetch } from './api.js';
import { isAdmin } from './auth.js';

let currentPage = 1;
let pageSize = 5;
let allData = [];
let filteredData = [];

export function renderList(data) {
    const container = document.getElementById('listContainer');
    if (!data || data.length === 0) {
        container.innerHTML = `<div class="empty">📭 Нет подписок</div>`;
        return;
    }

    // Получаем ID текущего пользователя
    let currentUserId = '';
    try {
        const userData = JSON.parse(localStorage.getItem('jwt_user') || '{}');
        currentUserId = userData.id || userData.user_id || '';
    } catch (e) {
        currentUserId = '';
    }

    const isUserAdmin = isAdmin();

    let html = `<table><thead><tr>
        <th>ID</th><th>Сервис</th><th>Цена</th><th>User ID</th><th>Начало</th><th>Окончание</th>
        <th>Действия</th>
    </tr></thead><tbody>`;

    data.forEach(s => {
        // Может ли пользователь редактировать эту подписку?
        const canEdit = isUserAdmin || (s.user_id === currentUserId);

        let actionsHtml = '';
        if (canEdit) {
            actionsHtml = `
                <button class="btn btn-outline" style="padding:4px 10px;font-size:12px;" 
                        onclick="window.openEditModal(${s.id})">✏️</button>
                <button class="btn btn-danger" style="padding:4px 10px;font-size:12px;" 
                        onclick="window.deleteSubscription(${s.id})">🗑️</button>
            `;
        } else {
            actionsHtml = `<span style="font-size:12px;color:#6b7280;">только чтение</span>`;
        }

        html += `<tr>
            <td><span class="badge">${s.id}</span></td>
            <td><strong>${escapeHtml(s.service_name)}</strong></td>
            <td>${s.price} ₽</td>
            <td style="font-size:13px;font-family:monospace;">${escapeHtml(s.user_id)}</td>
            <td>${s.start_date || '—'}</td>
            <td>${s.end_date || '—'}</td>
            <td>${actionsHtml}</td>
        </tr>`;
    });

    html += `</tbody></table>`;
    container.innerHTML = html;
}

function renderPagination() {
    const totalItems = filteredData.length;
    const totalPages = Math.ceil(totalItems / pageSize);
    const container = document.getElementById('paginationContainer');
    if (!container) return;

    if (totalPages <= 1) {
        container.innerHTML = '';
        return;
    }

    let html = `<div style="display:flex; gap:8px; align-items:center; margin-top:16px; flex-wrap:wrap;">`;
    html += `<button class="btn btn-outline btn-sm" onclick="window.goToPage(${currentPage - 1})" ${currentPage <= 1 ? 'disabled' : ''}>◀</button>`;
    html += `<span>Страница ${currentPage} из ${totalPages}</span>`;
    html += `<button class="btn btn-outline btn-sm" onclick="window.goToPage(${currentPage + 1})" ${currentPage >= totalPages ? 'disabled' : ''}>▶</button>`;
    html += `</div>`;
    container.innerHTML = html;
}

export function loadSubscriptions(userId = '', serviceName = '') {
    const container = document.getElementById('listContainer');
    container.innerHTML = '⏳ Загрузка...';

    let url = '/subscriptions';
    const params = new URLSearchParams();
    if (userId) params.append('user_id', userId);
    if (serviceName) params.append('service_name', serviceName);
    if (params.toString()) url += '?' + params.toString();

    apiFetch(url)
        .then(response => {
            const data = Array.isArray(response) ? response : [];
            allData = data;
            filteredData = data;
            currentPage = 1;
            renderPage();
        })
        .catch(err => {
            container.innerHTML = `<div class="empty">❌ Ошибка: ${err.message}</div>`;
            showToast(err.message, true);
        });
}

function renderPage() {
    const start = (currentPage - 1) * pageSize;
    const end = start + pageSize;
    const pageData = filteredData.slice(start, end);
    renderList(pageData);
    renderPagination();
}

window.goToPage = function(page) {
    const totalPages = Math.ceil(filteredData.length / pageSize);
    if (page < 1 || page > totalPages) return;
    currentPage = page;
    renderPage();
};

window.applyFiltersFromUI = function(userId = '', serviceName = '') {
    loadSubscriptions(userId, serviceName);
};

// ============================================================
// ПАГИНАЦИЯ: ИЗМЕНЕНИЕ КОЛИЧЕСТВА ЗАПИСЕЙ НА СТРАНИЦЕ
// ============================================================
document.addEventListener('DOMContentLoaded', () => {
    const select = document.getElementById('pageSizeSelect');
    if (select) {
        select.addEventListener('change', function() {
            pageSize = parseInt(this.value);
            currentPage = 1;
            loadSubscriptions(
                document.getElementById('filterUserId')?.value || '',
                document.getElementById('filterServiceName')?.value || ''
            );
        });
    }
});