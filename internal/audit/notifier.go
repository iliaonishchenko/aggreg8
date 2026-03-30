package audit

// Notifier определяет интерфейс рассылки аудит-событий всем зарегистрированным наблюдателям.
type Notifier interface {
	NotifyAll(event AuditEvent) []error
}

// NotifierService реализует Notifier, рассылая события всем зарегистрированным Observer.
type NotifierService struct {
	observers []Observer
}

// NewNotifier создаёт новый NotifierService без наблюдателей.
func NewNotifier() *NotifierService {
	return &NotifierService{
		observers: make([]Observer, 0),
	}
}

// Register добавляет наблюдателя в список рассылки.
func (n *NotifierService) Register(observer Observer) {
	n.observers = append(n.observers, observer)
}

// NotifyAll рассылает событие всем наблюдателям и возвращает список ошибок.
func (n *NotifierService) NotifyAll(event AuditEvent) []error {
	errors := make([]error, 0)

	for _, observer := range n.observers {
		err := observer.Notify(event)
		if err != nil {
			errors = append(errors, err)
		}
	}
	return errors
}
