// ============================================================
// ПАКЕТ: authentication
// ФАЙЛ: grpc_interceptor.go
// НАЗНАЧЕНИЕ: gRPC Interceptor для проверки JWT-токенов
// ============================================================
//
// ЧТО ТАКОЕ gRPC INTERCEPTOR?
// ============================================================
// Interceptor — это middleware для gRPC.
// Он перехватывает ВСЕ входящие gRPC-запросы ДО того,
// как они достигнут вашего сервиса.
//
// Аналогия:
//   - В REST это middleware (AuthMiddleware)
//   - В gRPC это interceptor
//
// ЗАЧЕМ НУЖЕН ЭТОТ INTERCEPTOR?
// ============================================================
// 1. Проверяет, есть ли токен в запросе
// 2. Проверяет, валидный ли токен (JWT)
// 3. Сохраняет данные пользователя в контекст
// 4. Если токен невалиден — возвращает ОШИБКУ
// 5. Если токен валиден — передаёт управление дальше
//
// ВАЖНО:
//   - Interceptor применяется КО ВСЕМ gRPC-методам
//   - Если метод не требует авторизации — можно сделать исключение
//   - Мы используем УНАРНЫЙ interceptor (для обычных RPC-методов)
// ============================================================

package authentication

import (
	"context"
	"strings"

	"Rest-user-agregator/pkg/logger"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// ============================================================
// AuthInterceptor — проверяет JWT-токен в gRPC-запросах
// ============================================================
// Это УНАРНЫЙ interceptor — применяется к методам с 1 запросом и 1 ответом.
// (Streaming interceptor нужен для потоковых методов — отдельно)
//
// ПАРАМЕТРЫ:
//   - ctx: контекст запроса (содержит metadata)
//   - req: запрос (данные от клиента)
//   - info: информация о вызываемом методе
//   - handler: функция, которая вызывает реальный метод
//
// ВОЗВРАЩАЕТ:
//   - ответ от реального метода
//   - ошибку, если токен невалиден
//
// АЛГОРИТМ:
//   1. Извлечь metadata из контекста
//   2. Найти заголовок Authorization
//   3. Проверить формат "Bearer <token>"
//   4. Валидировать JWT-токен
//   5. Сохранить данные пользователя в контекст
//   6. Передать управление дальше (handler)
// ============================================================
func AuthInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	logger.Debug("gRPC AuthInterceptor: received request for method: %s", info.FullMethod)
	// ============================================================
	// ШАГ 1: ИЗВЛЕКАЕМ METADATA ИЗ КОНТЕКСТА
	// ============================================================
	// metadata — это аналог HTTP-заголовков для gRPC.
	// Клиент передаёт их вместе с запросом.
	// Например: Authorization, User-Agent, Content-Type, и т.д.
	//
	// В Go это выглядит как map[string][]string
	// Например: {"authorization": ["Bearer eyJhbGci..."]}
	//
	// metadata.FromIncomingContext() — извлекает metadata из контекста.
	// Возвращает:
	//   - md: metadata (map)
	//   - ok: true если metadata есть, false если нет
	// ============================================================
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		// Если metadata нет — это не авторизованный запрос
		// Возвращаем ошибку Unauthenticated (аналог 401 Unauthorized)
		logger.Warn("gRPC AuthInterceptor: missing metadata")
		return nil, status.Error(codes.Unauthenticated, "missing metadata")
	}

	// ============================================================
	// ШАГ 2: ИЗВЛЕКАЕМ ТОКЕН ИЗ ЗАГОЛОВКА AUTHORIZATION
	// ============================================================
	// Клиент может передать заголовок в двух вариантах:
	//   - "authorization" (маленькие буквы) — чаще всего
	//   - "Authorization" (с большой буквы) — тоже работает
	//
	// Пробуем оба варианта.
	// ============================================================
	
	// Пробуем получить заголовок "authorization" (маленькие буквы)
	authHeader := md.Get("authorization")
	
	// Если не нашли — пробуем "Authorization" (с большой буквы)
	if len(authHeader) == 0 {
		authHeader = md.Get("Authorization")
	}
	
	// Если заголовка нет — запрос не авторизован
	if len(authHeader) == 0 {
		logger.Warn("gRPC AuthInterceptor: missing Authorization header")
		return nil, status.Error(codes.Unauthenticated, "authorization header required")
	}

	// ============================================================
	// ШАГ 3: ПРОВЕРЯЕМ ФОРМАТ "Bearer <token>"
	// ============================================================
	// Стандартный формат для JWT-токенов:
	//   Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
	//
	// Мы должны проверить:
	//   1. Начинается ли строка с "Bearer "
	//   2. Есть ли что-то после "Bearer "
	//
	// Если формат неверный — возвращаем ошибку.
	// ============================================================
	tokenString := authHeader[0]
	const prefix = "Bearer "
	
	// Проверяем, начинается ли токен с "Bearer "
	if !strings.HasPrefix(tokenString, prefix) {
		logger.Warn("gRPC AuthInterceptor: invalid Authorization header format")
		return nil, status.Error(codes.Unauthenticated, "invalid authorization format")
	}
	
	// Отрезаем префикс "Bearer " и получаем чистый токен
	tokenString = strings.TrimPrefix(tokenString, prefix)
logger.Debug("gRPC AuthInterceptor: token extracted, first 20 chars: %s...", tokenString[:20])
	// ============================================================
	// ШАГ 4: ВАЛИДИРУЕМ JWT-ТОКЕН
	// ============================================================
	// ValidateToken — функция из пакета authentication.
	// Она проверяет:
	//   1. Подпись токена (секрет из .env)
	//   2. Срок действия (не истёк ли)
	//   3. Формат токена (JWT)
	//
	// Возвращает:
	//   - claims: данные из токена (UserID, Email, Role)
	//   - error: ошибка, если токен невалиден
	// ============================================================
	claims, err := ValidateToken(tokenString)
	if err != nil {
		// Токен невалиден — логируем и возвращаем ошибку
		logger.Warn("gRPC AuthInterceptor: invalid token: %v", err)
		return nil, status.Error(codes.Unauthenticated, "invalid or expired token")
	}

	// ============================================================
	// ШАГ 5: СОХРАНЯЕМ ДАННЫЕ ПОЛЬЗОВАТЕЛЯ В КОНТЕКСТ
	// ============================================================
	// После того как токен проверен, мы должны сохранить данные
	// пользователя в контекст, чтобы методы могли их использовать.
	//
	// Это аналогично тому, как в REST мы сохраняли данные в контекст:
	//   ctx = context.WithValue(ctx, UserIDKey, claims.UserID)
	//
	// Ключи: UserIDKey, EmailKey, RoleKey — определены в middleware.go
	// Они используются в gRPC-методах для получения user_id и роли.
	// ============================================================
	ctx = context.WithValue(ctx, UserIDKey, claims.UserID)
	ctx = context.WithValue(ctx, EmailKey, claims.Email)
	ctx = context.WithValue(ctx, RoleKey, claims.Role)

	// Логируем успешную аутентификацию
	logger.Debug("gRPC AuthInterceptor: user authenticated: %s (role: %s)", claims.Email, claims.Role)

	// ============================================================
	// ШАГ 6: ПЕРЕДАЁМ УПРАВЛЕНИЕ ДАЛЬШЕ
	// ============================================================
	// handler — это реальный gRPC-метод, который должен обработать запрос.
	// Например: CreateSubscription, GetSubscriptions, и т.д.
	//
	// Мы передаём:
	//   - ctx: контекст с данными пользователя
	//   - req: запрос (данные от клиента)
	//
	// Если метод выполнится успешно — вернёт ответ.
	// Если метод вернёт ошибку — она будет передана клиенту.
	// ============================================================
	return handler(ctx, req)
}

// ============================================================
// 2. ОПЦИОНАЛЬНО: PUBLIC METHODS WHITELIST
// ============================================================
// Если есть методы, которые НЕ ТРЕБУЮТ авторизации,
// можно добавить их в белый список.
//
// Например:
//   - HealthCheck
//   - PublicInfo
//
// Для этого нужно изменить AuthInterceptor и добавить проверку:
//
// publicMethods := map[string]bool{
//     "/subscription.SubscriptionService/HealthCheck": true,
// }
//
// if publicMethods[info.FullMethod] {
//     // Пропускаем без проверки токена
//     return handler(ctx, req)
// }
//
// Сейчас ВСЕ методы требуют авторизацию.
// Если понадобится добавить публичные методы — раскомментируйте код ниже.
// ============================================================

// // publicMethods — список методов, которые НЕ требуют авторизации
// var publicMethods = map[string]bool{
// 	// "/subscription.SubscriptionService/HealthCheck": true,
// 	// "/subscription.SubscriptionService/PublicInfo": true,
// }
//
// // Использование в AuthInterceptor:
// // if publicMethods[info.FullMethod] {
// //     // Пропускаем без проверки токена
// //     return handler(ctx, req)
// // }

// ============================================================
// В РЕЗУЛЬТАТЕ:
// ============================================================
// 1. Все gRPC-запросы проходят через AuthInterceptor
// 2. Если токен отсутствует — возвращается ошибка Unauthenticated
// 3. Если токен невалиден — возвращается ошибка Unauthenticated
// 4. Если токен валиден — данные сохраняются в контекст
// 5. Методы получают user_id и роль из контекста
// 6. Код стал безопаснее и унифицирован с REST
// ============================================================