package audit

// AuditEvent содержит информацию о событии обновления метрик для аудита.
type AuditEvent struct {
	TS        int64    `json:"ts"`
	Metrics   []string `json:"metrics"`
	IPAddress string   `json:"ip_address"`
}

// Observer определяет интерфейс наблюдателя, получающего уведомления об аудит-событиях.
type Observer interface {
	Notify(event AuditEvent) error
}
