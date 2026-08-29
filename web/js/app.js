// ============================================================
// ГЛАВНЫЙ МОДУЛЬ ПРИЛОЖЕНИЯ (app.js)
// ============================================================

import { showToast, isValidDate } from './utils.js';
import { apiFetch, loadConfig, login, register } from './api.js';
import { logout, checkAuthAndRender, isAdmin, getUser } from './auth.js';
import { loadSubscriptions } from './components.js';

let editId = null;

// ============================================================
// ФУНКЦИЯ НАСТРОЙКИ МОДАЛЬНОГО ОКНА
// ============================================================
// ИЗМЕНЕНО: убрали isAdminFn
function setupEditModal(apiFetchFn, loadSubscriptionsFn, showToastFn, isValidDateFn) {
    let editId = null;
    let editUserId = null;

    // Открыть модалку
    window.openEditModal = function(id) {
        // ИЗМЕНЕНО: проверяем только авторизацию, не админа
        const token = localStorage.getItem('jwt_token');
        if (!token) {
            showToastFn('Необходимо войти в систему', true);
            return;
        }

        editId = id;
        const modal = document.getElementById('editModal');
        const title = document.getElementById('editModalTitle');
        
        title.textContent = `✏️ Редактирование подписки #${id}`;

        showToastFn('⏳ Загрузка...');
        apiFetchFn(`/subscriptions/${id}`)
            .then(sub => {
                editUserId = sub.user_id;
                document.getElementById('editStartDate').value = sub.start_date || '';
                document.getElementById('editEndDate').value = sub.end_date || '';
                modal.style.display = 'flex';
                showToastFn(`✏️ Редактирование подписки #${id}`);
            })
            .catch(err => {
                // ИЗМЕНЕНО: обработка ошибки 403
                if (err.message.includes('403') || err.message.includes('permission')) {
                    showToastFn('❌ Вы не можете редактировать эту подписку', true);
                } else {
                    showToastFn(`❌ Ошибка: ${err.message}`, true);
                }
            });
    };

    // Закрыть модалку
    function closeEditModal() {
        document.getElementById('editModal').style.display = 'none';
        document.getElementById('editForm').reset();
        editId = null;
        editUserId = null;
    }

    // Обработчик отправки формы редактирования
    function handleEditSubmit(e) {
        e.preventDefault();

        if (!editId) {
            showToastFn('❌ Ошибка: ID подписки не найден', true);
            return;
        }

        const startDate = document.getElementById('editStartDate').value.trim();
        const endDate = document.getElementById('editEndDate').value.trim() || '';

        if (!isValidDateFn(startDate)) {
            showToastFn('Дата начала должна быть в формате MM-YYYY', true);
            return;
        }
        if (endDate && !isValidDateFn(endDate)) {
            showToastFn('Дата окончания должна быть в формате MM-YYYY', true);
            return;
        }

        if (endDate) {
            const [m1, y1] = startDate.split('-').map(Number);
            const [m2, y2] = endDate.split('-').map(Number);
            if (y2 < y1 || (y2 === y1 && m2 < m1)) {
                showToastFn('Дата окончания не может быть раньше даты начала', true);
                return;
            }
        }

        const payload = {
            user_id: editUserId || '',
            start_date: startDate,
            end_date: endDate
        };
        
        console.log('📤 Отправляем payload (редактирование):', payload);

        apiFetchFn(`/subscriptions/${editId}`, {
            method: 'PUT',
            body: JSON.stringify(payload)
        })
        .then(() => {
            showToastFn(`✅ Подписка #${editId} обновлена`);
            closeEditModal();
            loadSubscriptionsFn();
        })
        .catch(err => {
            // ИЗМЕНЕНО: обработка ошибки 403
            if (err.message.includes('403') || err.message.includes('permission')) {
                showToastFn('❌ Вы не можете редактировать эту подписку', true);
            } else {
                showToastFn(`❌ Ошибка: ${err.message}`, true);
            }
        });
    }

    // Навешиваем обработчики
    document.getElementById('closeModalBtn')?.addEventListener('click', closeEditModal);
    document.getElementById('cancelEditModalBtn')?.addEventListener('click', closeEditModal);
    document.getElementById('editModal')?.addEventListener('click', function(e) {
        if (e.target === this) closeEditModal();
    });
    document.getElementById('editForm')?.addEventListener('submit', handleEditSubmit);

    return { openEditModal: window.openEditModal, closeEditModal };
}

// ============================================================
// ФУНКЦИЯ НАСТРОЙКИ УДАЛЕНИЯ
// ============================================================
// ИЗМЕНЕНО: убрали isAdminFn
function setupDeleteSubscription(apiFetchFn, loadSubscriptionsFn, showToastFn) {
    window.deleteSubscription = function(id) {
        // ИЗМЕНЕНО: проверяем только авторизацию, не админа
        const token = localStorage.getItem('jwt_token');
        if (!token) {
            showToastFn('Необходимо войти в систему', true);
            return;
        }

        if (!confirm(`Удалить подписку #${id}?`)) return;

        apiFetchFn(`/subscriptions/${id}`, { method: 'DELETE' })
            .then(() => {
                showToastFn(`✅ Подписка #${id} удалена`);
                loadSubscriptionsFn();
            })
            .catch(err => {
                // ИЗМЕНЕНО: обработка ошибки 403
                if (err.message.includes('403') || err.message.includes('permission') || err.message.includes('Permission denied')) {
                    showToastFn('❌ Вы не можете удалить эту подписку', true);
                } else {
                    showToastFn(`❌ Ошибка: ${err.message}`, true);
                }
            });
    };
}

// ============================================================
// ЗАПУСК ПРИЛОЖЕНИЯ
// ============================================================
document.addEventListener('DOMContentLoaded', async () => {
    // 1. Загрузка конфига
    try {
        await loadConfig();
        console.log('✅ Config loaded');
    } catch (error) {
        console.warn('⚠️ Config fallback:', error);
    }

    // 2. Проверка авторизации
    checkAuthAndRender();

    // 3. Настройка модального окна
    // ИЗМЕНЕНО: убрали isAdmin
    setupEditModal(
        apiFetch,
        loadSubscriptions,
        showToast,
        isValidDate
    );

    // 4. Настройка удаления
    // ИЗМЕНЕНО: убрали isAdmin
    setupDeleteSubscription(
        apiFetch,
        loadSubscriptions,
        showToast
    );

    // 5. Если пользователь уже вошёл
    if (localStorage.getItem('jwt_token')) {
        loadSubscriptions();
        await loadTemplates();

        if (isAdmin()) {
            document.getElementById('adminPanel').style.display = 'block';
        }
    }

    // ============================================================
    // ОСТАЛЬНЫЕ ОБРАБОТЧИКИ (БЕЗ ИЗМЕНЕНИЙ)
    // ============================================================

    // 5.1. ВХОД
    document.getElementById('loginBtn').addEventListener('click', () => {
    const email = document.getElementById('loginUsername').value.trim();
    const password = document.getElementById('loginPassword').value.trim();

    if (!email || !password) {
        showToast('Введите email и пароль', true);
        return;
    }

    login(email, password)
        .then(data => {
            console.log('🔍 Ответ от сервера:', data);
            
            // Сохраняем токен
            localStorage.setItem('jwt_token', data.token);
            
            // Пытаемся получить user_id из ответа
            let userId = data.user_id || data.id || '';
            
            // Если user_id нет в ответе — загружаем подписки
            if (!userId) {
                console.log('🔍 user_id нет в ответе, загружаем подписки...');
                
                // Загружаем подписки, чтобы получить user_id
                return apiFetch('/subscriptions?limit=1')
                    .then(subs => {
                        console.log('🔍 Получены подписки:', subs);
                        
                        // Если есть подписки — берём user_id из первой
                        if (subs && subs.length > 0 && subs[0].user_id) {
                            userId = subs[0].user_id;
                            console.log('✅ user_id получен из подписок:', userId);
                        } else {
                            // Если подписок нет — используем email (временно)
                            userId = email;
                            console.warn('⚠️ Подписок нет, используем email как ID:', userId);
                        }
                        
                        // Сохраняем пользователя
                        localStorage.setItem('jwt_user', JSON.stringify({
                            id: userId,
                            email: data.email,
                            role: data.role
                        }));
                        
                        console.log('✅ Пользователь сохранён:', JSON.parse(localStorage.getItem('jwt_user')));
                        
                        // Продолжаем инициализацию
                        finishLogin(data);
                    })
                    .catch(err => {
                        console.error('❌ Ошибка загрузки подписок:', err);
                        // Fallback: используем email
                        localStorage.setItem('jwt_user', JSON.stringify({
                            id: email,
                            email: data.email,
                            role: data.role
                        }));
                        finishLogin(data);
                    });
            }
            
            // Если user_id есть — сохраняем сразу
            localStorage.setItem('jwt_user', JSON.stringify({
                id: userId,
                email: data.email,
                role: data.role
            }));
            finishLogin(data);
        })
        .catch(err => showToast('❌ Ошибка: ' + err.message, true));
});

// Вспомогательная функция для завершения входа
function finishLogin(data) {
    showToast('✅ Добро пожаловать!');
    checkAuthAndRender();
    loadSubscriptions();
    loadTemplates();
    if (isAdmin()) {
        document.getElementById('adminPanel').style.display = 'block';
    }
}

    // 5.2. ВЫХОД
    document.getElementById('logoutBtn').addEventListener('click', logout);

    // 5.3. РЕГИСТРАЦИЯ
    document.getElementById('registerBtn')?.addEventListener('click', function() {
        const email = document.getElementById('regEmail').value.trim();
        const password = document.getElementById('regPassword').value.trim();

        if (!email || !password) {
            showToast('Введите email и пароль', true);
            return;
        }

        register(email, password, 'user')
            .then(() => {
                showToast('✅ Регистрация успешна! Теперь войдите.');
                document.getElementById('regEmail').value = '';
                document.getElementById('regPassword').value = '';
            })
            .catch(err => showToast('❌ Ошибка: ' + err.message, true));
    });

    // 5.4. СОЗДАНИЕ / ОБНОВЛЕНИЕ ПОДПИСКИ
    document.getElementById('createForm').addEventListener('submit', function(e) {
        e.preventDefault();

        const templateId = parseInt(document.getElementById('templateSelect').value);
        const startDate = document.getElementById('startDate').value.trim();
        const endDate = document.getElementById('endDate').value.trim() || '';

        if (!templateId) {
            showToast('Выберите шаблон', true);
            return;
        }
        if (!isValidDate(startDate)) {
            showToast('Дата начала должна быть в формате MM-YYYY', true);
            return;
        }
        if (endDate && !isValidDate(endDate)) {
            showToast('Дата окончания должна быть в формате MM-YYYY', true);
            return;
        }

        const payload = {
            template_id: templateId,
            start_date: startDate,
            end_date: endDate
        };
        console.log('📤 Отправляем payload (создание):', payload);

        const method = editId ? 'PUT' : 'POST';
        const url = editId ? `/subscriptions/${editId}` : '/subscriptions';

        apiFetch(url, { method, body: JSON.stringify(payload) })
            .then(data => {
                showToast(editId ? `✅ Подписка #${editId} обновлена` : `✅ Создано! ID: ${data.id}`);
                editId = null;
                document.getElementById('cancelEditBtn').style.display = 'none';
                document.getElementById('createForm').reset();
                loadSubscriptions();
            })
            .catch(err => showToast(`❌ Ошибка: ${err.message}`, true));
    });

    // 5.5. ОТМЕНА РЕДАКТИРОВАНИЯ
    document.getElementById('cancelEditBtn').addEventListener('click', function() {
        editId = null;
        this.style.display = 'none';
        document.getElementById('createForm').reset();
        showToast('Редактирование отменено');
    });

    // 5.6. ОБНОВЛЕНИЕ СПИСКА
    document.getElementById('refreshBtn').addEventListener('click', () => {
        loadSubscriptions();
    });

    // 5.7. РАСЧЁТ СУММАРНОЙ СТОИМОСТИ
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

        const url = `/subscriptions/total-cost?start_date=${startDate}&end_date=${endDate}`;

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

    // 5.8. УПРАВЛЕНИЕ ШАБЛОНАМИ
    const templatesBtn = document.getElementById('manageTemplatesBtn');
    if (templatesBtn) {
        templatesBtn.addEventListener('click', function() {
            window.location.href = '/templates.html';
        });
    }
});

// ============================================================
// ВСПОМОГАТЕЛЬНАЯ ФУНКЦИЯ: ЗАГРУЗКА ШАБЛОНОВ
// ============================================================
async function loadTemplates() {
    try {
        const templates = await apiFetch('/templates');
        console.log('Templates loaded:', templates);
        
        const select = document.getElementById('templateSelect');
        if (!select) return;

        select.innerHTML = '<option value="">Выберите шаблон</option>';

        templates.forEach(t => {
            const option = document.createElement('option');
            option.value = t.id;
            option.textContent = `${t.service_name} (${t.price} ₽)`;
            select.appendChild(option);
        });
    } catch (err) {
        console.error('Ошибка загрузки шаблонов:', err);
    }
}