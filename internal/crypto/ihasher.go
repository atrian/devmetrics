package crypto

// Hasher интерфейс для подписи метрик и проверки подлинности подписи
type Hasher interface {
	// Hash подписывает метрику. На вход получает метрику в виде строки и ключ, возвращает строку с подписью
	Hash(metric string, key string) string
	// Compare проверяет подпись метрики по предоставленному ключу
	Compare(hash string, metric string, key string) bool
}
