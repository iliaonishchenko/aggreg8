package audit_test

import (
	"fmt"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/iliaonishchenko/aggreg8/internal/audit"
	"github.com/iliaonishchenko/aggreg8/internal/audit/mocks"
	"github.com/stretchr/testify/assert"
)

func TestNotifierService_NotifyAll(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	obs1 := mocks.NewMockObserver(ctrl)
	obs2 := mocks.NewMockObserver(ctrl)

	event := audit.AuditEvent{
		TS:        12345,
		Metrics:   []string{"Alloc"},
		IPAddress: "10.0.0.1",
	}

	obs1.EXPECT().Notify(event).Return(nil)
	obs2.EXPECT().Notify(event).Return(nil)

	notifier := audit.NewNotifier(10)
	notifier.Register(obs1)
	notifier.Register(obs2)

	errs := notifier.NotifyAll(event)
	assert.Empty(t, errs)
}

func TestNotifierService_NotifyAll_CollectsErrors(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	obs1 := mocks.NewMockObserver(ctrl)
	obs2 := mocks.NewMockObserver(ctrl)
	obs3 := mocks.NewMockObserver(ctrl)

	event := audit.AuditEvent{TS: 1, Metrics: []string{"M"}, IPAddress: "1.2.3.4"}

	obs1.EXPECT().Notify(event).Return(fmt.Errorf("fail 1"))
	obs2.EXPECT().Notify(event).Return(nil)
	obs3.EXPECT().Notify(event).Return(fmt.Errorf("fail 2"))

	notifier := audit.NewNotifier(10)
	notifier.Register(obs1)
	notifier.Register(obs2)
	notifier.Register(obs3)

	errs := notifier.NotifyAll(event)
	assert.Len(t, errs, 2)
}

func TestNotifierService_NotifyAll_NoObservers(t *testing.T) {
	notifier := audit.NewNotifier(10)

	errs := notifier.NotifyAll(audit.AuditEvent{TS: 1, Metrics: []string{"M"}, IPAddress: "1.2.3.4"})
	assert.Empty(t, errs)
}
