import { showToast, escapeHtml } from './utils.js';
import { apiFetch, logout } from './api.js';
import { checkAuthAndRender, isAdmin, getUser } from './auth.js';

let editId = null;

document.addEventListener('DOMContentLoaded', async () => {
    // Проверка авторизации
    checkAuthAndRender();

    // Только админ имеет доступ
    if (!isAdmin()) {
        showToast('Доступ запрещён', true);
        setTimeout(() => window.location.href = '/', 1500);
        return;
    }

    // Показываем имя пользователя
    const user = getUser();
    document.getElementById('userNameDisplay').textContent = user?.email || 'Админ';

    // Загрузка шаблонов
    await loadTemplates();

    // ===== Кнопка "Назад" =====
    document.getElementById('backBtn').addEventListener('click', () => {
        window.location.href = '/';
    });

    // ===== Выход =====
    document.getElementById('logoutBtn').addEventListener('click', logout);

    // ===== Обновить =====
    document.getElementById('refreshBtn').addEventListener('click', loadTemplates);

    // ===== Создание / Обновление =====
    document.getElementById('templateForm').addEventListener('submit', async function(e) {
        e.preventDefault();

        const serviceName = document.getElementById('serviceName').value.trim();
        const price = parseInt(document.getElementById('price').value);

        if (!serviceName || isNaN(price) || price < 0) {
            showToast('Заполните все поля корректно', true);
            return;
        }

        try {
            let result;
            if (editId) {
                result = await apiFetch(`/admin/templates/${editId}`, {
                    method: 'PUT',
                    body: JSON.stringify({ service_name: serviceName, price })
                });
                showToast('✅ Шаблон обновлён');
            } else {
                result = await apiFetch('/admin/templates', {
                    method: 'POST',
                    body: JSON.stringify({ service_name: serviceName, price })
                });
                showToast('✅ Шаблон создан');
            }

            editId = null;
            document.getElementById('cancelEditBtn').style.display = 'none';
            document.getElementById('templateForm').reset();
            document.getElementById('formTitle').textContent = '➕ Новый шаблон';
            await loadTemplates();
        } catch (err) {
            showToast(`❌ Ошибка: ${err.message}`, true);
        }
    });

    // ===== Отмена редактирования =====
    document.getElementById('cancelEditBtn').addEventListener('click', function() {
        editId = null;
        this.style.display = 'none';
        document.getElementById('templateForm').reset();
        document.getElementById('formTitle').textContent = '➕ Новый шаблон';
        showToast('Редактирование отменено');
    });
});

// ===== Загрузка шаблонов =====
async function loadTemplates() {
    try {
        const templates = await apiFetch('/templates');
        const container = document.getElementById('listContainer');

        if (!templates || templates.length === 0) {
            container.innerHTML = '<div class="empty">📭 Нет шаблонов</div>';
            return;
        }

        let html = `<table><thead><tr>
            <th>ID</th>
            <th>Название</th>
            <th>Цена</th>
            <th>Действия</th>
        </tr></thead><tbody>`;

        templates.forEach(t => {
            html += `<tr>
                <td><span class="badge">${t.id}</span></td>
                <td><td><strong>${escapeHtml(t.service_name)}</strong></td>
                <td>${t.price} ₽</td>
                <td>
                    <button class="btn btn-outline btn-sm" onclick="editTemplate(${t.id})">✏️</button>
                    <button class="btn btn-danger btn-sm" onclick="deleteTemplate(${t.id})">🗑️</button>
                </td>
            </tr>`;
        });

        html += `</tbody></table>`;
        container.innerHTML = html;
    } catch (err) {
        showToast('❌ Ошибка загрузки шаблонов', true);
        console.error(err);
    }
}

// ===== Редактирование =====
window.editTemplate = async function(id) {
    try {
        const template = await apiFetch(`/templates/${id}`);
        document.getElementById('serviceName').value = template.service_name;
        document.getElementById('price').value = template.price;
        document.getElementById('formTitle').textContent = '✏️ Редактирование шаблона';
        document.getElementById('cancelEditBtn').style.display = 'inline-block';
        editId = id;
        showToast(`Редактирование шаблона #${id}`);
    } catch (err) {
        showToast(`❌ Ошибка: ${err.message}`, true);
    }
};

// ===== Удаление =====
window.deleteTemplate = async function(id) {
    if (!confirm(`Удалить шаблон #${id}?`)) return;

    try {
        await apiFetch(`/admin/templates/${id}`, { method: 'DELETE' });
        showToast(`✅ Шаблон #${id} удалён`);
        await loadTemplates();
    } catch (err) {
        showToast(`❌ Ошибка: ${err.message}`, true);
    }
};