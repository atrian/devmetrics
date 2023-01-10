package dto

// Metrics используется для передачи данных о метрике между слоями приложения и для маршаллинга/анмаршаллинга JSON
type Metrics struct {
	ID    string   `json:"id"`              // ID имя метрики
	MType string   `json:"type"`            // MType параметр, принимающий значение gauge или counter
	Delta *int64   `json:"delta,omitempty"` // Delta значение метрики в случае передачи counter
	Value *float64 `json:"value,omitempty"` // Value значение метрики в случае передачи gauge
	Hash  string   `json:"hash,omitempty"`  // Hash значение хеш-функции - подпись для проверки подлинности метрики
}
