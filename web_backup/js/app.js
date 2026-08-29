// ============================================================
// ГЛАВНЫЙ МОДУЛЬ ПРИЛОЖЕНИЯ (app.js)
// ============================================================
// Этот файл — "дирижёр" всего приложения.
// Он:
//   1. Ждёт загрузки страницы (DOMContentLoaded)
//   2. Настраивает кнопки и формы
//   3. Обрабатывает действия пользователя (вход, создание, редактирование)
//   4. Связывает интерфейс с бэкендом через api.js
// ============================================================

// ============================================================
// 1. ИМПОРТЫ (подключение других модулей)
// ============================================================
// import ... from ... — загружает функции из других файлов.
// Это как "взять инструменты из соседнего ящика".

// showToast — показать всплывающее сообщение (зелёное/красное)
// isValidDate — проверить, что дата в формате MM-YYYY
import { showToast, isValidDate } from './utils.js';

// apiFetch — универсальная функция для запросов к серверу
// loadConfig — загрузить адрес сервера (бэкенда)
// login, register — функции для входа и регистрации
import { apiFetch, loadConfig, login, register } from './api.js';

// logout — выйти из системы
// checkAuthAndRender — проверить, авторизован ли пользователь, и показать нужный интерфейс
// isAdmin — true, если пользователь администратор
// getUser — получить данные пользователя из localStorage
import { logout, checkAuthAndRender, isAdmin, getUser } from './auth.js';

// loadSubscriptions — загрузить список подписок с сервера и отрисовать таблицу
import { loadSubscriptions } from './components.js';

// ============================================================
// 2. ГЛОБАЛЬНЫЕ ПЕРЕМЕННЫЕ
// ============================================================
// let — переменная, которую можно менять.
// editId — хранит ID подписки, которую редактируем.
// Если null — значит создаём новую.
let editId = null;

// ============================================================
// 3. ЗАГРУЗКА ШАБЛОНОВ В ВЫПАДАЮЩИЙ СПИСОК
// ============================================================
// Эта функция запрашивает с сервера список всех шаблонов
// и заполняет ими выпадающий список <select id="templateSelect">
async function loadTemplates() {
    // try ... catch — перехват ошибок (если сервер не ответит)
    try {
        // 1. Запрашиваем шаблоны с сервера (GET /api/templates)
        // await — ждём ответа от сервера
        const templates = await apiFetch('/templates');
        console.log('Templates loaded:', templates);  // ← добавить
        
        // 2. Находим выпадающий список на странице по id
        const select = document.getElementById('templateSelect');
        if (!select) return; // если элемента нет — выходим

        // 3. Очищаем список, оставляя только первый пустой пункт
        // innerHTML — весь HTML-код внутри элемента
        select.innerHTML = '<option value="">Выберите шаблон</option>';

        // 4. Перебираем все шаблоны и добавляем их в список
        // forEach — цикл по массиву
        templates.forEach(t => {
            // createElement — создаём новый HTML-элемент <option>
            const option = document.createElement('option');
            option.value = t.id;   // значение, которое отправится на сервер
            option.textContent = `${t.service_name} (${t.price} ₽)`; // то, что видит пользователь
            select.appendChild(option); // добавляем в конец списка
        });
    } catch (err) {
        // Если ошибка — пишем в консоль (не показываем пользователю)
        console.error('Ошибка загрузки шаблонов:', err);
    }
}

// ============================================================
// 4. ЗАПУСК ПРИЛОЖЕНИЯ ПОСЛЕ ЗАГРУЗКИ СТРАНИЦЫ
// ============================================================
// DOMContentLoaded — событие, которое срабатывает, когда HTML-код
// полностью загружен и браузер построил DOM-дерево.
// Это гарантирует, что все элементы на странице уже существуют.
document.addEventListener('DOMContentLoaded', async () => {
//document.getElementById('registerForm').style.display = 'none';
    // 4.1. ЗАГРУЗКА КОНФИГА (адрес бэкенда)
    // loadConfig() — запрашивает у сервера его адрес (порт, хост)
    // Это нужно, чтобы фронтенд знал, куда отправлять запросы.
    try {
        await loadConfig();
        console.log('✅ Config loaded');
    } catch (error) {
        // Если сервер не ответил — используем localhost:8080 (запасной вариант)
        console.warn('⚠️ Config fallback:', error);
    }

    // 4.2. ПРОВЕРКА АВТОРИЗАЦИИ
    // checkAuthAndRender() — смотрит, есть ли в localStorage JWT-токен.
    // Если есть — показывает основной интерфейс (mainApp).
    // Если нет — показывает форму входа (loginForm).
    checkAuthAndRender();

    // 4.3. ЕСЛИ ПОЛЬЗОВАТЕЛЬ УЖЕ ВОШЁЛ
    if (localStorage.getItem('jwt_token')) {
        // Загружаем список подписок
        loadSubscriptions();
        // Загружаем шаблоны в выпадающий список
        await loadTemplates();

        // Если пользователь — администратор, показываем кнопку "Управление шаблонами"
        if (isAdmin()) {
            document.getElementById('adminPanel').style.display = 'block';
        }
    }

    // ============================================================
    // 5. ОБРАБОТЧИКИ СОБЫТИЙ (что происходит при кликах)
    // ============================================================
    // addEventListener — назначает функцию, которая выполнится
    // при наступлении события (click, submit, change...)

    // 5.1. ВХОД
    // При клике на кнопку "Войти" (loginBtn) выполняем:
    document.getElementById('loginBtn').addEventListener('click', () => {
        // Берём текст из полей ввода
        const email = document.getElementById('loginUsername').value.trim();
        const password = document.getElementById('loginPassword').value.trim();

        // Проверяем, что поля не пустые
        if (!email || !password) {
            showToast('Введите email и пароль', true);
            return;
        }

        // Отправляем запрос на сервер (login из api.js)
        login(email, password)
            .then(data => {
                // Если успешно — сохраняем токен и данные пользователя в localStorage
                localStorage.setItem('jwt_token', data.token);
                localStorage.setItem('jwt_user', JSON.stringify({
                    email: data.email,
                    role: data.role
                }));
                showToast('✅ Добро пожаловать!');
                // Перерисовываем интерфейс
                checkAuthAndRender();
                loadSubscriptions();
                loadTemplates();
                if (isAdmin()) {
                    document.getElementById('adminPanel').style.display = 'block';
                }
            })
            .catch(err => showToast('❌ Ошибка: ' + err.message, true));
    });

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

        // Отправляем запрос на регистрацию
        register(email, password, 'user')
            .then(() => {
                showToast('✅ Регистрация успешна! Теперь войдите.');
                // Очищаем поля формы
                document.getElementById('regEmail').value = '';
                document.getElementById('regPassword').value = '';
            })
            .catch(err => showToast('❌ Ошибка: ' + err.message, true));
    });

    // ============================================================
    // 5.4. СОЗДАНИЕ / ОБНОВЛЕНИЕ ПОДПИСКИ
    // ============================================================
    // submit — событие отправки формы (при нажатии Enter или кнопки submit)
    document.getElementById('createForm').addEventListener('submit', function(e) {
        // e.preventDefault() — отключаем стандартную отправку формы
        // (чтобы страница не перезагружалась)
        e.preventDefault();

        // Берём значения из полей формы
        const templateId = parseInt(document.getElementById('templateSelect').value);
        const startDate = document.getElementById('startDate').value.trim();
        const endDate = document.getElementById('endDate').value.trim() || '';

        // Валидация (проверка корректности)
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

        // Формируем данные для отправки на сервер
        // Важно: поля service_name, price, user_id — НЕ передаём,
        // потому что бэкенд берёт их из шаблона и из контекста.
        const payload = {
            template_id: templateId,
            start_date: startDate,
            end_date: endDate
        };

        // Определяем метод и URL в зависимости от того,
        // редактируем мы существующую подписку (editId !== null)
        // или создаём новую.
        const method = editId ? 'PUT' : 'POST';
        const url = editId ? `/subscriptions/${editId}` : '/subscriptions';

        // Отправляем запрос на сервер
        apiFetch(url, { method, body: JSON.stringify(payload) })
            .then(data => {
                showToast(editId ? `✅ Подписка #${editId} обновлена` : `✅ Создано! ID: ${data.id}`);
                // Сбрасываем состояние редактирования
                editId = null;
                document.getElementById('cancelEditBtn').style.display = 'none';
                document.getElementById('createForm').reset();
                // Перезагружаем список подписок
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

    // ============================================================
    // 5.6. РЕДАКТИРОВАНИЕ ПОДПИСКИ
    // ============================================================
    // Эта функция вызывается из таблицы (components.js)
    // при клике на кнопку "✏️" (редактировать).
    window.editSubscription = function(id) {
        // Только администратор может редактировать
        if (!isAdmin()) {
            showToast('Только администратор может редактировать', true);
            return;
        }

        // Запрашиваем данные подписки с сервера
        apiFetch(`/subscriptions/${id}`)
            .then(sub => {
                editId = id;
                // Заполняем форму только датами
                document.getElementById('startDate').value = sub.start_date || '';
                document.getElementById('endDate').value = sub.end_date || '';
                document.getElementById('cancelEditBtn').style.display = 'inline-block';
                showToast(`Редактирование подписки #${id}`);
            })
            .catch(err => showToast(`❌ Ошибка: ${err.message}`, true));
    };

    // ============================================================
    // 5.7. УДАЛЕНИЕ ПОДПИСКИ
    // ============================================================
    window.deleteSubscription = function(id) {
        if (!isAdmin()) {
            showToast('Только администратор может удалять', true);
            return;
        }
        // confirm — показывает модальное окно с вопросом
        if (!confirm(`Удалить подписку #${id}?`)) return;

        apiFetch(`/subscriptions/${id}`, { method: 'DELETE' })
            .then(() => {
                showToast(`✅ Подписка #${id} удалена`);
                loadSubscriptions();
            })
            .catch(err => showToast(`❌ Ошибка: ${err.message}`, true));
    };

    // ============================================================
    // 5.8. ОБНОВЛЕНИЕ СПИСКА (кнопка "🔄 Обновить")
    // ============================================================
    document.getElementById('refreshBtn').addEventListener('click', () => {
        loadSubscriptions();
    });

    // ============================================================
    // 5.9. РАСЧЁТ СУММАРНОЙ СТОИМОСТИ (TOTAL-COST)
    // ============================================================
    document.getElementById('calcTotalBtn').addEventListener('click', function() {
        const startDate = document.getElementById('totalStartDate').value.trim();
        const endDate = document.getElementById('totalEndDate').value.trim();

        // Проверяем, что даты введены
        if (!startDate || !endDate) {
            showToast('Введите обе даты (MM-YYYY)', true);
            return;
        }
        if (!isValidDate(startDate) || !isValidDate(endDate)) {
            showToast('Даты должны быть в формате MM-YYYY', true);
            return;
        }

        // Формируем URL для запроса
        // Важно: НЕ передаём user_id и service_name,
        // потому что бэкенд сам определяет,
        // какие подписки показывать (все для админа, только свои для юзера).
        const url = `/subscriptions/total-cost?start_date=${startDate}&end_date=${endDate}`;

        // Отправляем запрос
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

    // ============================================================
    // 5.10. УПРАВЛЕНИЕ ШАБЛОНАМИ (кнопка для админа)
    // ============================================================
    const templatesBtn = document.getElementById('manageTemplatesBtn');
    if (templatesBtn) {
        templatesBtn.addEventListener('click', function() {
            window.location.href = '/templates.html';
        });
    }

}); // КОНЕЦ DOMContentLoaded