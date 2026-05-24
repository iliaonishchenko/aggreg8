package agent

import (
	"context"
	"github.com/golang/mock/gomock"
	"github.com/iliaonishchenko/aggreg8/internal/agent/mocks"
	"testing"
	"time"
)

func TestRun(t *testing.T) {
	t.Run("collects on poll interval", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		mockCollector := mocks.NewMockCollectorService(ctrl)
		mockSender := mocks.NewMockSenderService(ctrl)
		agent := NewAgent(mockCollector, mockSender, 1, 2, 2)
		ctx, cancel := context.WithCancel(context.Background())

		mockCollector.EXPECT().CollectSystem().MinTimes(2)
		mockCollector.EXPECT().CollectRuntime().MinTimes(2)
		mockCollector.EXPECT().GetMetrics().AnyTimes()

		go agent.Run(ctx)

		time.Sleep(3 * time.Second)
		cancel()
		time.Sleep(100 * time.Millisecond)
	})

	t.Run("sends on report interval", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		mockCollector := mocks.NewMockCollectorService(ctrl)
		mockSender := mocks.NewMockSenderService(ctrl)
		agent := NewAgent(mockCollector, mockSender, 1, 2, 2)
		ctx, cancel := context.WithCancel(context.Background())

		mockCollector.EXPECT().CollectSystem().AnyTimes()
		mockCollector.EXPECT().CollectRuntime().AnyTimes()
		mockCollector.EXPECT().GetMetrics().AnyTimes()
		mockSender.EXPECT().SendJSONWithRetries(gomock.Any()).AnyTimes()

		go agent.Run(ctx)

		time.Sleep(3 * time.Second)
		cancel()
		time.Sleep(100 * time.Millisecond)
	})
}
