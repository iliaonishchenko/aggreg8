package audit

// AuditEvent содержит информацию о событии обновления метрик для аудита.
type AuditEvent struct {
	TS        int64    `json:"ts"`
	Metrics   []string `json:"metrics"`
	IPAddress string   `json:"ip_address"`
}
