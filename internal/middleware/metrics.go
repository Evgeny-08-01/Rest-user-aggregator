// ============================================================
// ФАЙЛ: internal/middleware/metrics.go
// ============================================================
// НАЗНАЧЕНИЕ: оборачивает хендлеры для сбора HTTP-метрик
//
// ЧТО ЗДЕСЬ ПРОИСХОДИТ:
//   1. Замеряем время начала обработки запроса
//   2. Вызываем оригинальный хендлер
//   3. Замеряем время окончания
//   4. Сохраняем метрики: количество запросов + время ответа
//
// ЭТОТ MIDDLEWARE ОБОРАЧИВАЕТСЯ ВОКРУГ КАЖДОГО ХЕНДЛЕРА:
//   MetricsMiddleware(handler) → возвращает функцию, которая:
//   - считает время
//   - вызывает handler
//   - записывает метрики
// ============================================================

package middleware

import (
	"net/http"
	"time"

	"Rest-user-agregator/internal/authentication"
	"Rest-user-agregator/internal/metrics"
	"Rest-user-agregator/pkg/logger"
)

// ============================================================
// MetricsMiddleware — обёртка для сбора метрик
// ============================================================
// Принимает:   http.HandlerFunc (любой хендлер)
// Возвращает:  http.HandlerFunc (хендлер с метриками)
//
// КАК РАБОТАЕТ:
//   1. Запоминаем время перед вызовом хендлера
//   2. Создаём обёртку для ResponseWriter, чтобы перехватить статус-код
//   3. Вызываем хендлер
//   4. Считаем длительность
//   5. Сохраняем метрики в Prometheus
// ============================================================
func MetricsMiddleware(next http.HandlerFunc) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        // 1. ЗАПОМИНАЕМ ВРЕМЯ СТАРТА
        //    time.Now() — текущее время
        //    start — переменная, в которой хранится время начала запроса
        start := time.Now()

        // 2. СОЗДАЁМ ОБЁРТКУ ДЛЯ RESPONSE WRITER
        //    Нам нужно перехватить статус-код, который вернёт хендлер.
        //    Стандартный ResponseWriter не даёт прочитать статус,
        //    поэтому мы создаём свою структуру responseWriter.
        rw := &responseWriter{
            ResponseWriter: w, // оригинальный writer
            statusCode:     200,  // статус по умолчанию, если хендлер не вызовет WriteHeader()
        }

        // 3. ВЫЗЫВАЕМ ОРИГИНАЛЬНЫЙ ХЕНДЛЕР
        //    Передаём ему нашу обёртку вместо оригинального w.
        //    Хендлер будет писать ответ через неё, а мы сможем
        //    перехватить статус-код.
       logger.Debug("MetricsMiddleware: before next, user_id=%v, email=%v, role=%v",
    r.Context().Value(authentication.UserIDKey),
    r.Context().Value(authentication.EmailKey),
    r.Context().Value(authentication.RoleKey),

) 
        next(rw, r)

        // 4. СЧИТАЕМ ДЛИТЕЛЬНОСТЬ
        //    time.Since(start) — разница между текущим временем и start
        //    .Seconds() — переводим в секунды (тип float64)
        duration := time.Since(start).Seconds()

        // 5. СОХРАНЯЕМ МЕТРИКИ
        //    http_requests_total — увеличиваем счётчик на 1
        //    Лейблы: метод, путь, статус-код (текстом)
        metrics.HttpRequestsTotal.WithLabelValues(
            r.Method,                   // GET, POST, PUT, DELETE
            r.URL.Path,                // /api/subscriptions, /api/login, ...
            http.StatusText(rw.statusCode), // "200 OK", "400 Bad Request", ...
        ).Inc()

         //    http_request_duration_seconds — записываем длительность
         //    Лейблы: метод, путь
        // 1. Берём гистограмму (счётчик времени)
        // 2. С помощью WithLabelValues говорим: "это время для метода GET и пути /api/subscriptions"
        // 3. Записываем в неё длительность (duration)
metrics.HttpRequestDuration.WithLabelValues(r.Method, r.URL.Path).Observe(duration)
    }
}

// ============================================================
// responseWriter — обёртка для перехвата статус-кода
// ============================================================
// Зачем:      стандартный http.ResponseWriter не даёт прочитать статус
// Что делаем: сохраняем статус в поле statusCode при вызове WriteHeader
// ============================================================
type responseWriter struct {
    http.ResponseWriter // встраиваем стандартный ResponseWriter
    statusCode int      // сюда сохраняем статус-код
}

// ============================================================
// WriteHeader — перехватывает установку статус-кода
// ============================================================
// Когда хендлер вызывает w.WriteHeader(200), мы:
//   1. Сохраняем статус в rw.statusCode
//   2. Вызываем оригинальный WriteHeader
// ============================================================
func (rw *responseWriter) WriteHeader(code int) {
    rw.statusCode = code
    rw.ResponseWriter.WriteHeader(code)
}