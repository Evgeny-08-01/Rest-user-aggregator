// ============================================================
// КОМПОНЕНТЫ (пагинация на фронтенде)
// ============================================================

import { escapeHtml, showToast } from './utils.js';
import { apiFetch } from './api.js';
import { isAdmin } from './auth.js';

let currentPage = 1;
let pageSize = 5; 
let allData = [];          // ← все данные, загруженные с бэкенда
let filteredData = [];     // ← данные после фильтрации

export function renderList(data) {
    const container = document.getElementById('listContainer');
    if (!data || data.length === 0) {
        container.innerHTML = `<div class="empty">📭 Нет подписок</div>`;
        return;
    }

    const isUserAdmin = isAdmin();

    let html = `<table><thead><tr>
        <th>ID</th><th>Сервис</th><th>Цена</th><th>User ID</th><th>Начало</th><th>Окончание</th>
        <th>Действия</th>
    </tr></thead><tbody>`;

    data.forEach(s => {
        html += `<tr>
            <td><span class="badge">${s.id}</span></td>
            <td><strong>${escapeHtml(s.service_name)}</strong></td>
            <td>${s.price} ₽</td>
            <td style="font-size:13px;font-family:monospace;">${escapeHtml(s.user_id)}</td>
            <td>${s.start_date || '—'}</td>
            <td>${s.end_date || '—'}</td>
            <td>
                ${isUserAdmin ? `
                    <button class="btn btn-outline" style="padding:4px 10px;font-size:12px;" onclick="window.editSubscription(${s.id})">✏️</button>
                    <button class="btn btn-danger" style="padding:4px 10px;font-size:12px;" onclick="window.deleteSubscription(${s.id})">🗑️</button>
                ` : `
                    <span style="font-size:12px;color:#6b7280;">только чтение</span>
                `}
            </td>
        </tr>`;
    });

    html += `</tbody></table>`;
    container.innerHTML = html;

    renderPagination();
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

    // Загружаем ВСЕ данные
    apiFetch('/subscriptions')
        .then(response => {
            // Если пришёл массив — используем его
            if (Array.isArray(response)) {
                allData = response;
            } else {
                allData = response.data || [];
            }

            // Применяем фильтры
            applyFilters(userId, serviceName);
        })
        .catch(err => {
            container.innerHTML = `<div class="empty">❌ Ошибка: ${err.message}</div>`;
            showToast(err.message, true);
        });
}

function applyFilters(userId, serviceName) {
    let filtered = allData;

    if (userId) {
        filtered = filtered.filter(s => s.user_id.includes(userId));
    }
    if (serviceName) {
        filtered = filtered.filter(s => s.service_name.toLowerCase().includes(serviceName.toLowerCase()));
    }

    filteredData = filtered;

    // Сброс на первую страницу
    currentPage = 1;
    renderPage();
}

function renderPage() {
    const start = (currentPage - 1) * pageSize;
    const end = start + pageSize;
    const pageData = filteredData.slice(start, end);
    renderList(pageData);
}

// Глобальные функции для кнопок
window.goToPage = function(page) {
    const totalPages = Math.ceil(filteredData.length / pageSize);
    if (page < 1 || page > totalPages) return;
    currentPage = page;
    renderPage();
};

// Обновляем фильтры извне
window.applyFiltersFromUI = function() {
    const userId = document.getElementById('filterUserId')?.value || '';
    const serviceName = document.getElementById('filterServiceName')?.value || '';
    loadSubscriptions(userId, serviceName);
};

// ============================================================
// ИЗМЕНЕНИЕ КОЛИЧЕСТВА ЗАПИСЕЙ НА СТРАНИЦЕ (ПАГИНАЦИЯ)
// ============================================================
// Этот код позволяет пользователю выбирать, сколько подписок
// показывать на одной странице (5, 10, 25, 50 или 100).
//
// Когда пользователь меняет значение в выпадающем списке,
// страница автоматически обновляется и показывает нужное количество записей.
//
// Это улучшает пользовательский опыт (UX) и даёт гибкость.
// ============================================================

// ============================================================
// 1. СЛУШАЕМ СОБЫТИЕ ЗАГРУЗКИ СТРАНИЦЫ (DOMContentLoaded)
// ============================================================
// DOMContentLoaded — событие, которое срабатывает, когда HTML-документ
// полностью загружен и готов к работе с JavaScript.
//
// Это гарантирует, что все элементы на странице уже существуют,
// и мы можем безопасно обращаться к ним через document.getElementById.
// ============================================================
document.addEventListener('DOMContentLoaded', () => {

    // ============================================================
    // 2. НАХОДИМ ВЫПАДАЮЩИЙ СПИСОК (SELECT) ПО ID
    // ============================================================
    // document.getElementById('pageSizeSelect') — находит элемент <select>
    // с id="pageSizeSelect". Этот элемент мы добавили в index.html.
    //
    // Если элемент найден — сохраняем его в переменную select.
    // Если не найден — select будет null (но мы это проверяем дальше).
    // ============================================================
    const select = document.getElementById('pageSizeSelect');

    // ============================================================
    // 3. ПРОВЕРЯЕМ, ЧТО ЭЛЕМЕНТ СУЩЕСТВУЕТ (if (select))
    // ============================================================
    // Если select === null (элемент не найден), то код внутри if
    // не выполнится. Это защита от ошибок, если мы случайно
    // удалим выпадающий список из HTML или переименуем id.
    // ============================================================
    if (select) {

        // ============================================================
        // 4. НАВЕШИВАЕМ ОБРАБОТЧИК СОБЫТИЯ 'change' НА SELECT
        // ============================================================
        // addEventListener('change', function() { ... }) — это способ
        // сказать браузеру: «Когда пользователь меняет значение в этом
        // выпадающем списке — выполни эту функцию».
        //
        // Событие 'change' срабатывает, когда пользователь выбирает
        // новый вариант в выпадающем списке (например, вместо 10 выбрал 25).
        // ============================================================
        select.addEventListener('change', function() {

            // ============================================================
            // 5. ПОЛУЧАЕМ НОВОЕ ЗНАЧЕНИЕ И ПРЕОБРАЗУЕМ В ЧИСЛО
            // ============================================================
            // this.value — содержит текущее выбранное значение из <option>.
            // Например, если выбрано "25", то this.value === "25".
            //
            // parseInt(this.value) — превращает строку "25" в число 25.
            // Это нужно, потому что pageSize должно быть числом (для расчётов).
            // ============================================================
            pageSize = parseInt(this.value);

            // ============================================================
            // 6. СБРАСЫВАЕМ НА ПЕРВУЮ СТРАНИЦУ
            // ============================================================
            // currentPage = 1 — когда пользователь меняет количество записей
            // на странице, мы автоматически возвращаем его на первую страницу.
            //
            // Почему? Если мы были на странице 3 и выбрали 100 записей,
            // то страница 3 может оказаться пустой (записей может не хватить).
            // Чтобы этого избежать — сбрасываем на первую страницу.
            // ============================================================
            currentPage = 1;

            // ============================================================
            // 7. ПЕРЕЗАГРУЖАЕМ ТАБЛИЦУ С УЧЁТОМ ФИЛЬТРОВ
            // ============================================================
            // loadSubscriptions(...) — функция, которая загружает данные
            // с сервера и обновляет таблицу.
            //
            // Мы передаём текущие значения фильтров (user_id и service_name),
            // чтобы сохранить применённые фильтры и не сбрасывать их.
            //
            // document.getElementById('filterUserId')?.value — получает значение
            // из поля фильтра по User ID. Если поля нет — возвращает пустую строку.
            // ?. — это «опциональная цепочка» (optional chaining), защита от ошибок.
            // ============================================================
            loadSubscriptions(
                document.getElementById('filterUserId')?.value || '',
                document.getElementById('filterServiceName')?.value || ''
            );
        });
    }
});