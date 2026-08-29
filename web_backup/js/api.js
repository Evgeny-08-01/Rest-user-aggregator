// ============================================================
// api.js — работа с сервером (с моками)
// ============================================================
// Этот файл — главный мост между фронтендом и твоим Go-бэкендом.
// Все запросы к серверу проходят через этот файл.
//
// КОГДА ИСПОЛЬЗУЕТСЯ:
// - Когда нужно загрузить список подписок (GET /subscriptions)
// - Когда нужно создать подписку (POST /subscriptions)
// - Когда нужно войти (POST /login)
// - Когда нужно рассчитать сумму (GET /subscriptions/total-cost)
// - И т.д.
//
// ЧТО ЗДЕСЬ ПРОИСХОДИТ:
// 1. Формируется URL (адрес сервера + путь).
// 2. Добавляется JWT-токен (если пользователь авторизован).
// 3. Отправляется запрос на сервер.
// 4. Обрабатывается ответ или ошибка.
// ============================================================

// ============================================================
// ИМПОРТЫ
// ============================================================
// Подключаем функцию showToast из файла utils.js.
// Она нужна, чтобы показывать всплывающие уведомления об ошибках.
import { showToast } from './utils.js';

// ============================================================
// !!! ВАЖНО: УБРАН ЦИКЛИЧЕСКИЙ ИМПОРТ !!!
// ============================================================
// Раньше здесь было: import { loadConfig } from './api.js';
// Это создавало циклическую зависимость (файл импортирует сам себя),
// из-за чего скрипт ломался и кнопка логина не работала.
// 
// Теперь loadConfig объявлен ниже как обычная функция,
// и мы не импортируем её из самого себя.
// ============================================================

// ============================================================
// БАЗОВЫЙ АДРЕС СЕРВЕРА (загружается динамически)
// ============================================================
// Вместо жёстко зашитого localhost:8080 мы загружаем адрес
// через отдельный эндпоинт /api/config. Это позволяет:
//   1. Менять порт без пересборки фронтенда.
//   2. Запускать бэкенд на любом хосте.
//   3. Использовать один и тот же фронтенд для разных окружений.
//
// АДРЕС ЗАГРУЖАЕТСЯ АСИНХРОННО:
//   - При старте приложения вызывается loadConfig()
//   - Она запрашивает /api/config и сохраняет адрес
//   - Все запросы идут на этот адрес
// ============================================================

// Переменная для хранения адреса (изначально пустая)
let API_BASE = '';

// Флаг, который показывает, загружен ли уже конфиг.
// Нужен чтобы не делать повторные запросы к /api/config.
let configLoaded = false;

// ============================================================
// ФУНКЦИЯ ЗАГРУЗКИ КОНФИГА (вызывается один раз при старте)
// ============================================================
// Эта функция запрашивает у сервера его адрес и сохраняет его.
// Если сервер не ответил — использует fallback (http://localhost:8080/api).
// 
// !!! ВАЖНО: Это теперь обычная функция, а не импорт из самого себя !!!
// ============================================================
export async function loadConfig() {
    // Если конфиг уже загружен и адрес есть — просто возвращаем его.
    // Это предотвращает лишние запросы к серверу.
    if (configLoaded && API_BASE) {
        console.log('ℹ️ Config already loaded, using cached value:', API_BASE);
        return API_BASE;
    }

    try {
        // 1. Отправляем запрос на /api/config
        // Используем относительный путь, потому что фронтенд и бэкенд
        // могут быть на одном хосте (например, localhost:8080).
        // Если фронтенд отдаётся через другой порт, то запрос всё равно
        // пойдёт на тот же хост и порт, откуда загружена страница.
        const response = await fetch('/api/config');
        
        // 2. Проверяем, что ответ успешный
        if (!response.ok) {
            throw new Error(`HTTP ${response.status}`);
        }
        
        // 3. Превращаем JSON в объект
        const config = await response.json();
        
        // 4. Сохраняем адрес бэкенда
        API_BASE = config.apiBase;
        configLoaded = true; // Помечаем, что конфиг загружен
        
        console.log('✅ Config loaded. API_BASE:', API_BASE);
        return API_BASE;
    } catch (error) {
        // 5. Если ошибка — используем fallback
        // Определяем порт из текущего URL страницы
        // Если порт не указан (стандартный 80 или 443), используем 8080
        const port = window.location.port || '8080';
        API_BASE = `http://localhost:${port}/api`;
        configLoaded = true;
        
        console.error('❌ Failed to load config:', error);
        console.warn('⚠️ Using fallback API_BASE:', API_BASE);
        return API_BASE;
    }
}

// ============================================================
// РЕЖИМ РАБОТЫ (моки или реальный бэкенд)
// ============================================================
// true  → данные берутся из моков (фейковые данные внутри этого файла)
// false → данные берутся с реального бэкенда (Go-сервер)
const USE_MOCK = false;  // ← ВЫКЛЮЧЕНЫ МОКИ - используем реальный бэкенд

// ============================================================
// МОКОВЫЕ ДАННЫЕ (используются, если USE_MOCK = true)
// ============================================================
// Здесь хранятся фейковые подписки, чтобы фронтенд работал без бэкенда.
// Это как "игрушечная" база данных в памяти браузера.
let mockSubscriptions = [
    { id: 1, service_name: 'Яндекс Плюс', price: 400, user_id: '550e8400-e29b-41d4-a716-446655440000', start_date: '07-2025', end_date: '12-2025' },
    { id: 2, service_name: 'Spotify', price: 250, user_id: '550e8400-e29b-41d4-a716-446655440001', start_date: '01-2025', end_date: '' },
];
// Счётчик для генерации новых ID в моках (чтобы каждый раз был новый номер)
let mockIdCounter = 3;

// ============================================================
// ГЛАВНАЯ ФУНКЦИЯ ДЛЯ ЗАПРОСОВ (apiFetch)
// ============================================================
// Эта функция вызывается из других файлов (app.js, components.js).
// Она принимает:
//   - path → куда идём (например, '/subscriptions')
//   - opts → настройки запроса (метод, тело, заголовки)
//
// Внутри она:
//   1. Загружает конфиг, если он ещё не загружен.
//   2. Добавляет JWT-токен в заголовок (если есть).
//   3. Если USE_MOCK = true → вызывает мок-функцию.
//   4. Если USE_MOCK = false → отправляет реальный fetch на сервер.
// ============================================================
export async function apiFetch(path, opts = {}) {
    // ===== 0. Загружаем конфиг, если ещё не загружен =====
    // Это гарантирует, что API_BASE всегда будет установлен
    // перед отправкой любого запроса.
    if (!API_BASE) {
        await loadConfig();
    }
    
    // ===== 1. Берём токен из localStorage =====
    // localStorage — это хранилище в браузере, где мы сохраняем JWT-токен после входа.
    // Если токена нет — значит, пользователь не авторизован.
    const token = localStorage.getItem('jwt_token');

    // ===== 2. Формируем заголовки запроса =====
    // Заголовки — это служебная информация, которую сервер читает перед обработкой.
    const headers = {
        // Говорим серверу, что мы отправляем JSON-данные (Content-Type).
        'Content-Type': 'application/json',
        
        // Если токен есть — добавляем его в заголовок Authorization.
        // Это стандартный способ передачи JWT: Bearer <токен>.
        ...(token ? { 'Authorization': `Bearer ${token}` } : {}),
        
        // Если в opts переданы свои заголовки — добавляем их (например, для логина).
        ...(opts.headers || {}),
    };

    // ===== 3. Выбор режима (мок или реальный бэкенд) =====
    // Если USE_MOCK = true — используем фейковые данные.
    // Если false — отправляем реальный запрос на сервер.
    if (USE_MOCK) {
        return mockFetch(path, opts);
    }

    // ===== 4. Формируем полный URL =====
    // Склеиваем базовый адрес и путь.
    const url = `${API_BASE}${path}`;
    
    // Логируем запрос в консоль для отладки.
    // Это помогает видеть, куда реально отправляются запросы.
    console.log(`🚀 [apiFetch] ${opts.method || 'GET'} ${url}`);

    // ===== 5. Реальный запрос на сервер (fetch) =====
    // fetch — это встроенная функция браузера для HTTP-запросов.
    return fetch(url, {
        ...opts,          // передаём метод (GET, POST, ...) и тело запроса
        headers,          // передаём заголовки (с токеном)
    })
    // ===== 6. Обработка ответа от сервера =====
    .then(async res => {
        // Пытаемся превратить ответ в JSON.
        // Если ответ пустой — data будет null.
        const data = await res.json().catch(() => null);
        
        // Если статус ответа НЕ 200–299 (res.ok = false) — выбрасываем ошибку.
        if (!res.ok) {
            // Берём сообщение об ошибке из ответа сервера (data.error или data.message).
            // Если сервер не прислал сообщение — пишем стандартное.
            const msg = data?.error || data?.message || `Ошибка ${res.status}`;
            throw new Error(msg);
        }
        
        // Если всё хорошо — возвращаем данные (JSON-объект).
        return data;
    });
}

// ============================================================
// МОК-ФУНКЦИЯ (используется при USE_MOCK = true)
// ============================================================
// Эта функция полностью имитирует работу бэкенда.
// Она не отправляет реальные запросы, а просто работает с массивом mockSubscriptions.
// Это нужно, чтобы фронтенд можно было тестировать без запуска Go-сервера.
function mockFetch(path, opts) {
    
    // Возвращаем Promise — так же, как настоящий fetch.
    // Promise — это обещание, что результат будет получен (успешно или с ошибкой).
    return new Promise((resolve, reject) => {
        
        // ===== 1. Разбираем метод и тело запроса =====
        // method — GET, POST, PUT, DELETE.
        // body — данные, которые отправляем (например, при создании подписки).
        const method = opts.method || 'GET';
        const body = opts.body ? JSON.parse(opts.body) : null;

        // ===== 2. Проверка авторизации (кроме /login) =====
        // Если путь НЕ /login (то есть мы не пытаемся войти),
        // И в localStorage нет токена — значит, пользователь не авторизован.
        // Возвращаем ошибку 401 (Unauthorized).
        if (!path.includes('/login') && !localStorage.getItem('jwt_token')) {
            reject(new Error('Unauthorized'));
            return;
        }

        // ============================================================
        // 1. LOGIN (POST /login)
        // ============================================================
        // Эмуляция входа в систему.
        if (method === 'POST' && path === '/login') {
            const { username, password } = body;

            // Проверка: логин и пароль должны совпадать (admin/admin или user/user).
            if (username === 'admin' && password === 'admin') {
                // Генерируем фейковый JWT-токен (просто закодированная строка).
                // В реальном бэкенде токен создаётся с помощью секретного ключа.
                const token = btoa(JSON.stringify({ username, role: 'admin', exp: Date.now() + 3600000 }));
                resolve({ token, username, role: 'admin' });
                return;
            }
            if (username === 'user' && password === 'user') {
                const token = btoa(JSON.stringify({ username, role: 'user', exp: Date.now() + 3600000 }));
                resolve({ token, username, role: 'user' });
                return;
            }

            // Если логин/пароль не совпадают — ошибка.
            reject(new Error('Неверный логин или пароль'));
            return;
        }

        // ============================================================
        // 2. GET /subscriptions (список всех подписок)
        // ============================================================
        if (method === 'GET' && path === '/subscriptions') {
            // Разбираем параметры из URL (limit, offset, user_id, service_name).
            const urlParams = new URLSearchParams(path.split('?')[1]);
            const limit = parseInt(urlParams.get('limit')) || 10;
            const offset = parseInt(urlParams.get('offset')) || 0;
            const userId = urlParams.get('user_id') || '';
            const serviceName = urlParams.get('service_name') || '';

            // Фильтруем моковые данные по user_id и service_name.
            let filtered = mockSubscriptions;
            if (userId) {
                filtered = filtered.filter(s => s.user_id.includes(userId));
            }
            if (serviceName) {
                filtered = filtered.filter(s => s.service_name.toLowerCase().includes(serviceName.toLowerCase()));
            }

            // Применяем пагинацию (берём только нужный кусок массива).
            const paginated = filtered.slice(offset, offset + limit);

            // Возвращаем объект с данными, общим количеством и параметрами пагинации.
            resolve({
                data: paginated,
                total: filtered.length,
                limit,
                offset,
            });
            return;
        }

        // ============================================================
        // 3. GET /subscriptions/{id} (получение одной подписки)
        // ============================================================
        // Регулярное выражение ищет в пути цифры после /subscriptions/.
        const getMatch = path.match(/^\/subscriptions\/(\d+)$/);
        if (method === 'GET' && getMatch) {
            const id = parseInt(getMatch[1]);
            // Ищем подписку с нужным ID в моковом массиве.
            const sub = mockSubscriptions.find(s => s.id === id);
            if (!sub) {
                reject(new Error('Not found'));
                return;
            }
            resolve(sub);
            return;
        }

        // ============================================================
        // 4. POST /subscriptions (создание новой подписки)
        // ============================================================
        if (method === 'POST' && path === '/subscriptions') {
            // Добавляем новый ID (счётчик увеличиваем).
            const newSub = { ...body, id: mockIdCounter++ };
            // Добавляем в массив моковых данных.
            mockSubscriptions.push(newSub);
            // Возвращаем ID созданной подписки (как настоящий бэкенд).
            resolve({ id: newSub.id });
            return;
        }

        // ============================================================
        // 5. PUT /subscriptions/{id} (обновление подписки)
        // ============================================================
        if (method === 'PUT' && getMatch) {
            const id = parseInt(getMatch[1]);
            const index = mockSubscriptions.findIndex(s => s.id === id);
            if (index === -1) {
                reject(new Error('Not found'));
                return;
            }
            // Обновляем подписку: берём старые данные и заменяем их новыми.
            mockSubscriptions[index] = { ...mockSubscriptions[index], ...body };
            resolve({ status: 'updated' });
            return;
        }

        // ============================================================
        // 6. DELETE /subscriptions/{id} (удаление подписки)
        // ============================================================
        if (method === 'DELETE' && getMatch) {
            const id = parseInt(getMatch[1]);
            const index = mockSubscriptions.findIndex(s => s.id === id);
            if (index === -1) {
                reject(new Error('Not found'));
                return;
            }
            // Удаляем подписку из массива.
            mockSubscriptions.splice(index, 1);
            resolve({ status: 'deleted' });
            return;
        }

        // ============================================================
        // 7. GET /subscriptions/total-cost (суммарная стоимость)
        // ============================================================
        if (method === 'GET' && path.includes('/subscriptions/total-cost')) {
            // Считаем сумму всех подписок (в реальном бэкенде это делает SQL).
            const total = mockSubscriptions.reduce((sum, sub) => sum + sub.price, 0);
            resolve({ total });
            return;
        }

        // ============================================================
        // Если ни один маршрут не подошёл — ошибка.
        // ============================================================
        reject(new Error('Not implemented'));
    });
}

// ============================================================
// ЛОГИН (через apiFetch, для реального бэкенда)
// ============================================================
// ============================================================
// АВТОРИЗАЦИЯ ПОЛЬЗОВАТЕЛЯ (ВХОД В СИСТЕМУ)
// ============================================================
// Эта функция отправляет запрос на вход и получает JWT-токен.
// 
// ПАРАМЕТРЫ:
//   - email    (string)  → Email пользователя (обязательно)
//   - password (string)  → Пароль пользователя (обязательно)
//
// ВОЗВРАЩАЕТ:
//   - Promise с данными от сервера
//   - При успехе: { token: "jwt_token", email: "...", role: "user" }
//   - При ошибке: { error: "..." }
//
// ИСПОЛЬЗОВАНИЕ:
//   import { login } from './api.js';
//   login('user@example.com', '123456')
//     .then(data => {
//         localStorage.setItem('jwt_token', data.token);
//     })
//     .catch(err => console.error('Ошибка входа:', err));
// ============================================================
export function login(email, password) {
    // 1. Вызываем apiFetch с методом POST
    return apiFetch('/login', {
        method: 'POST',                          // HTTP метод
        body: JSON.stringify({                   // Превращаем объект в JSON-строку
            email: email,                        // Email пользователя
            password: password,                  // Пароль (будет проверен на бэкенде)
        }),
    });
}

// ============================================================
// ВЫХОД
// ============================================================
// Удаляет токен и данные пользователя из localStorage.
// Перезагружает страницу, чтобы сбросить состояние приложения.
// ============================================================
export function logout() {
    localStorage.removeItem('jwt_token');
    localStorage.removeItem('jwt_user');
    window.location.reload();
}

// ============================================================
// ЗАПРОСЫ К БЭКЕНДУ
// ============================================================
// Эти функции являются удобными обёртками над apiFetch.
// Они вызываются из app.js и components.js.
// ============================================================

// GET /subscriptions — получить список всех подписок
export function getSubscriptions() {
    return apiFetch('/subscriptions');
}

// POST /subscriptions — создать новую подписку
export function createSubscription(data) {
    // data должен содержать: { template_id, start_date, end_date }
    return apiFetch('/subscriptions', {
        method: 'POST',
        body: JSON.stringify({
            template_id: data.template_id,
            start_date: data.start_date,
            end_date: data.end_date || ''
        }),
    });
}

// PUT /subscriptions/{id} — обновить существующую подписку
export function updateSubscription(id, startDate, endDate) {
    return apiFetch(`/subscriptions/${id}`, {
        method: 'PUT',
        body: JSON.stringify({
            start_date: startDate,
            end_date: endDate || ''
        }),
    });
}

// DELETE /subscriptions/{id} — удалить подписку
export function deleteSubscription(id) {
    return apiFetch(`/subscriptions/${id}`, {
        method: 'DELETE',
    });
}

// GET /subscriptions/total-cost — получить суммарную стоимость подписок за период
// Параметры:
//   - startDate: дата начала (MM-YYYY)
//   - endDate: дата окончания (MM-YYYY)
//   - userId: ID пользователя (опционально, для фильтрации)
//   - serviceName: название сервиса (опционально, для фильтрации)
export function getTotalCost(startDate, endDate) {
    return apiFetch(`/subscriptions/total-cost?start_date=${startDate}&end_date=${endDate}`);
}
// ============================================================
// РЕГИСТРАЦИЯ НОВОГО ПОЛЬЗОВАТЕЛЯ
// ============================================================
// Эта функция отправляет запрос на создание нового пользователя.
// 
// ПАРАМЕТРЫ:
//   - email    (string)  → Email пользователя (обязательно)
//   - password (string)  → Пароль пользователя (обязательно)
//   - role     (string)  → Роль: 'user' или 'admin' (по умолчанию 'user')
//
// ВОЗВРАЩАЕТ:
//   - Promise с данными ответа от сервера
//   - При успехе: { message: "User registered successfully" }
//   - При ошибке: { error: "..." }
//
// ИСПОЛЬЗОВАНИЕ:
//   import { register } from './api.js';
//   register('user@example.com', '123456', 'user')
//     .then(data => console.log('Успех:', data))
//     .catch(err => console.error('Ошибка:', err));
// ============================================================
export function register(email, password, role = 'user') {
    // 1. Вызываем apiFetch с методом POST
    //    - Первый аргумент: '/register' — эндпоинт на бэкенде
    //    - Второй аргумент: объект с настройками запроса
    return apiFetch('/register', {
        method: 'POST',                          // HTTP метод
        body: JSON.stringify({                   // Превращаем объект в JSON-строку
            email: email,                        // Email пользователя
            password: password,                  // Пароль (будет захеширован на бэкенде)
            role: role,                          // Роль (по умолчанию 'user')
        }),
    });
}