package audit

import "sync"

// Notifier определяет интерфейс рассылки аудит-событий всем зарегистрированным наблюдателям.
type Notifier interface {
	NotifyAll(event AuditEvent) []error
}

// Observer определяет интерфейс наблюдателя, получающего уведомления об аудит-событиях.
type Observer interface {
	Notify(event AuditEvent) error
}

// NotifierService реализует Notifier, рассылая события всем зарегистрированным Observer.
type NotifierService struct {
	mu        sync.RWMutex
	observers []Observer
	sem       chan struct{}
}

// NewNotifier создаёт новый NotifierService без наблюдателей.
func NewNotifier(maxWorkers int) *NotifierService {
	return &NotifierService{
		observers: make([]Observer, 0),
		sem:       make(chan struct{}, maxWorkers),
	}
}

// Register добавляет наблюдателя в список рассылки.
func (n *NotifierService) Register(observer Observer) {
	n.mu.Lock()
	n.observers = append(n.observers, observer)
	n.mu.Unlock()
}

// NotifyAll рассылает событие всем наблюдателям и возвращает список ошибок.
func (n *NotifierService) NotifyAll(event AuditEvent) []error {
	n.mu.RLock()
	observers := n.observers
	n.mu.RUnlock()

	var (
		mu   sync.Mutex
		errs []error
		wg   sync.WaitGroup
	)

	for _, observer := range observers {
		n.sem <- struct{}{}
		wg.Add(1)
		go func(o Observer) {
			defer wg.Done()
			defer func() { <-n.sem }()
			if err := o.Notify(event); err != nil {
				mu.Lock()
				errs = append(errs, err)
				mu.Unlock()
			}
		}(observer)
	}

	wg.Wait()
	return errs
}
