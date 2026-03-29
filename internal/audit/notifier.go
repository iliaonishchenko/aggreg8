package audit

type Notifier interface {
	NotifyAll(event AuditEvent) []error
}

type NotifierService struct {
	observers []Observer
}

func NewNotifier() *NotifierService {
	return &NotifierService{
		observers: make([]Observer, 0),
	}
}

func (n *NotifierService) Register(observer Observer) {
	n.observers = append(n.observers, observer)
}

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
