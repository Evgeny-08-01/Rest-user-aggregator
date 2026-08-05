// ============================================================
// ГЛАВНЫЙ МОДУЛЬ
// ============================================================

import { showToast, isValidUUID, isValidDate } from './utils.js';
// !!! ВАЖНО: Импортируем loadConfig из api.js !!!
import { apiFetch, loadConfig } from './api.js';
import { login, logout, checkAuthAndRender, isAdmin, getUser } from './auth.js';
import { loadSubscriptions, renderList } from './components.js';

let editId = null;

// ===== ИНИЦИАЛИЗАЦИЯ =====
document.addEventListener('DOMContentLoaded', async () => {
    // ============================================================
    // !!! ВАЖНО: Загружаем конфиг ПЕРВЫМ делом !!!
    // ============================================================
    // Это гарантирует, что API_BASE будет загружен до того,
    // как пользователь нажмёт кнопку "Войти" или выполнит любой другой запрос.
    // 
    // Без этого apiFetch не будет знать, куда отправлять запросы,
    // и кнопка логина не будет работать.
    // ============================================================
    try {
        await loadConfig();
        console.log('✅ Config loaded successfully');
    } catch (error) {
        console.warn('⚠️ Config loading failed, using fallback:', error);
        // Даже если конфиг не загрузился, apiFetch использует fallback
        // Поэтому приложение продолжит работать
    }

    // Проверяем авторизацию и показываем нужный интерфейс
    checkAuthAndRender();

    // Если пользователь авторизован — загружаем подписки
    if (localStorage.getItem('jwt_token')) {
        loadSubscriptions();
    }

    // ===== ВХОД =====
    document.getElementById('loginBtn').addEventListener('click', () => {
        const username = document.getElementById('loginUsername').value.trim();
        const password = document.getElementById('loginPassword').value.trim();
        if (!username || !password) {
            showToast('Введите логин и пароль', true);
            return;
        }
        login(username, password)
            .then(() => {
                showToast(`✅ Добро пожаловать, ${username}!`);
                checkAuthAndRender();
                loadSubscriptions();
            })
            .catch(() => {}); // Ошибка уже обрабатывается в login()
    });

    // ===== ВЫХОД =====
    document.getElementById('logoutBtn').addEventListener('click', logout);

    // ===== СОЗДАНИЕ / ОБНОВЛЕНИЕ =====
    document.getElementById('createForm').addEventListener('submit', function(e) {
        e.preventDefault();

        const payload = {
            service_name: document.getElementById('serviceName').value.trim(),
            price: parseInt(document.getElementById('price').value),
            user_id: document.getElementById('userId').value.trim(),
            start_date: document.getElementById('startDate').value.trim(),
            end_date: document.getElementById('endDate').value.trim() || '',
        };

        if (!payload.service_name || payload.price === undefined || payload.price === null || payload.price < 0 || !payload.user_id || !payload.start_date) {
            showToast('Заполните все обязательные поля (цена должна быть >= 0)', true);
            return;
        }
        if (!isValidUUID(payload.user_id)) {
            showToast('Неверный формат UUID', true);
            return;
        }
        if (!isValidDate(payload.start_date) || (payload.end_date && !isValidDate(payload.end_date))) {
            showToast('Дата должна быть в формате MM-YYYY', true);
            return;
        }

        const method = editId ? 'PUT' : 'POST';
        const url = editId ? `/subscriptions/${editId}` : '/subscriptions';

        apiFetch(url, { method, body: JSON.stringify(payload) })
            .then(data => {
                showToast(editId ? `✅ Подписка #${editId} обновлена` : `✅ Создано! ID: ${data.id}`);
                editId = null;
                document.getElementById('cancelEditBtn').style.display = 'none';
                document.getElementById('createForm').reset();
                loadSubscriptions(
                    document.getElementById('filterUserId')?.value || '',
                    document.getElementById('filterServiceName')?.value || ''
                );
            })
            .catch(err => showToast(`❌ Ошибка: ${err.message}`, true));
    });

    // ===== ОТМЕНА РЕДАКТИРОВАНИЯ =====
    document.getElementById('cancelEditBtn').addEventListener('click', function() {
        editId = null;
        this.style.display = 'none';
        document.getElementById('createForm').reset();
        showToast('Редактирование отменено');
    });

    // ===== РЕДАКТИРОВАНИЕ =====
    window.editSubscription = function(id) {
        if (!isAdmin()) {
            showToast('Только администратор может редактировать', true);
            return;
        }
        apiFetch(`/subscriptions/${id}`)
            .then(sub => {
                editId = id;
                document.getElementById('serviceName').value = sub.service_name || '';
                document.getElementById('price').value = sub.price || '';
                document.getElementById('userId').value = sub.user_id || '';
                document.getElementById('startDate').value = sub.start_date || '';
                document.getElementById('endDate').value = sub.end_date || '';
                document.getElementById('cancelEditBtn').style.display = 'inline-block';
                showToast(`Редактирование подписки #${id}`);
            })
            .catch(err => showToast(`❌ Ошибка: ${err.message}`, true));
    };

    // ===== УДАЛЕНИЕ =====
    window.deleteSubscription = function(id) {
        if (!isAdmin()) {
            showToast('Только администратор может удалять', true);
            return;
        }
        if (!confirm(`Удалить подписку #${id}?`)) return;
        apiFetch(`/subscriptions/${id}`, { method: 'DELETE' })
            .then(() => {
                showToast(`✅ Подписка #${id} удалена`);
                loadSubscriptions(
                    document.getElementById('filterUserId')?.value || '',
                    document.getElementById('filterServiceName')?.value || ''
                );
            })
            .catch(err => showToast(`❌ Ошибка: ${err.message}`, true));
    };

    // ===== ОБНОВЛЕНИЕ СПИСКА =====
    document.getElementById('refreshBtn').addEventListener('click', () => {
        loadSubscriptions(
            document.getElementById('filterUserId')?.value || '',
            document.getElementById('filterServiceName')?.value || ''
        );
    });

    // ===== ФИЛЬТРЫ =====
    document.getElementById('applyFiltersBtn').addEventListener('click', () => {
        const userId = document.getElementById('filterUserId').value.trim();
        const serviceName = document.getElementById('filterServiceName').value.trim();
        // Сброс пагинации на первую страницу
        window.goToPage(1);
        loadSubscriptions(userId, serviceName);
    });

    document.getElementById('resetFiltersBtn').addEventListener('click', () => {
        document.getElementById('filterUserId').value = '';
        document.getElementById('filterServiceName').value = '';
        window.goToPage(1);
        loadSubscriptions('', '');
        showToast('Фильтры сброшены');
    });

    // ===== ТОТАЛ КОСТ С УЧЁТОМ ФИЛЬТРОВ =====
    document.getElementById('calcTotalBtn').addEventListener('click', function() {
        const startDate = document.getElementById('totalStartDate').value.trim();
        const endDate = document.getElementById('totalEndDate').value.trim();

        if (!startDate || !endDate) {
            showToast('Введите обе даты (MM-YYYY)', true);
            return;
        }
        if (!isValidDate(startDate) || !isValidDate(endDate)) {
            showToast('Даты должны быть в формате MM-YYYY', true);
            return;
        }

        // Берём текущие фильтры
        const userId = document.getElementById('filterUserId')?.value.trim() || '';
        const serviceName = document.getElementById('filterServiceName')?.value.trim() || '';

        // Формируем URL с параметрами
        let url = `/subscriptions/total-cost?start_date=${startDate}&end_date=${endDate}`;
        if (userId) {
            url += `&user_id=${encodeURIComponent(userId)}`;
        }
        if (serviceName) {
            url += `&service_name=${encodeURIComponent(serviceName)}`;
        }

        apiFetch(url)
            .then(data => {
                document.getElementById('totalResult').innerHTML = `💰 Итого: <span style="color:#16a34a;">${data.total} ₽</span>`;
                showToast(`✅ Суммарная стоимость: ${data.total} ₽`);
            })
            .catch(err => {
                document.getElementById('totalResult').innerHTML = `❌ Ошибка: ${err.message}`;
                showToast(`❌ Ошибка: ${err.message}`, true);
            });
    });
});