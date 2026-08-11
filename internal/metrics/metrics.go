// ============================================================
// ФАЙЛ: internal/metrics/metrics.go
// ============================================================
// НАЗНАЧЕНИЕ: объявляет HTTP-метрики для Prometheus
//
// ЧТО ЗДЕСЬ ПРОИСХОДИТ:
//   1. Создаём счётчик http_requests_total — считает количество запросов
//   2. Создаём гистограмму http_request_duration_seconds — измеряет время ответа
//
// ЭТИ МЕТРИКИ ПОЗВОЛЯТ:
//   - Увидеть RPS (запросов в секунду) по каждому эндпоинту
//   - Увидеть задержки (средние, 95-й процентиль)
//   - Отслеживать количество ошибок (статусы 4xx, 5xx)
// ============================================================

package metrics

import (
    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promauto"
)

// ============================================================
// 1. СЧЁТЧИК ЗАПРОСОВ (http_requests_total)
// ============================================================
// Что это:   счётчик, который увеличивается на 1 при каждом запросе
// Зачем:     считать RPS (запросов в секунду)
// Лейблы:    method (GET/POST/PUT/DELETE)
//            path   (/api/subscriptions, /api/login, ...)
//            status (200, 400, 401, 500, ...)
//
// В Grafana запрос для RPS будет выглядеть так:
//   rate(http_requests_total[1m])
// ============================================================
var HttpRequestsTotal = promauto.NewCounterVec(
    prometheus.CounterOpts{
        Name: "http_requests_total",
        Help: "Total number of HTTP requests",
    },
    []string{"method", "path", "status"},
)

// ============================================================
// 2. ГИСТОГРАММА ЗАДЕРЖЕК (http_request_duration_seconds)
// ============================================================
// Что это:   измеряет время ответа в секундах
// Зачем:     видеть, сколько времени обрабатывается запрос
// Лейблы:    method (GET/POST/PUT/DELETE)
//            path   (/api/subscriptions, /api/login, ...)
// Buckets:   заранее заданные интервалы для группировки:
//            1ms, 5ms, 10ms, 25ms, 50ms, 100ms, 250ms, 500ms, 1s, 2.5s, 5s, 10s
//
// В Grafana запрос для 95-го процентиля задержки:
//   histogram_quantile(0.95, rate(http_request_duration_seconds_bucket[1m]))
// ============================================================
var HttpRequestDuration = promauto.NewHistogramVec(
    prometheus.HistogramOpts{
        Name:    "http_request_duration_seconds",
        Help:    "HTTP request duration in seconds",
        Buckets: []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10},
    },
    []string{"method", "path"},
)