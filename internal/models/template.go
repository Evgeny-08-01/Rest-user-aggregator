package models

// Template — шаблон подписки, создаваемый админом
// Используется пользователями для создания своих подписок
// Поля:
//   - ID: уникальный идентификатор шаблона
//   - ServiceName: название сервиса (например, "Яндекс Плюс")
//   - Price: цена подписки в рублях (целое число, >= 0)
type Template struct {
	ID          int    `json:"id"`
	ServiceName string `json:"service_name"`
	Price       int    `json:"price"`
}
