package agent

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/iliaonishchenko/aggreg8/internal/logger"
	models "github.com/iliaonishchenko/aggreg8/internal/model"
	pb "github.com/iliaonishchenko/aggreg8/internal/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// realIPMetadataKey — ключ метаданных gRPC, в котором агент передаёт свой IP-адрес.
const realIPMetadataKey = "x-real-ip"

// GRPCSender отправляет метрики на сервер по gRPC, реализуя SenderService.
type GRPCSender struct {
	conn    *grpc.ClientConn
	client  pb.MetricsClient
	localIP string
}

// NewGRPCSender создаёт gRPC-клиент для отправки метрик на указанный адрес.
func NewGRPCSender(target, localIP string) (*GRPCSender, error) {
	conn, err := grpc.NewClient(target, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("ошибка создания gRPC-клиента: %w", err)
	}
	return &GRPCSender{
		conn:    conn,
		client:  pb.NewMetricsClient(conn),
		localIP: localIP,
	}, nil
}

// SendJSONWithRetries отправляет батч метрик через UpdateMetrics RPC с автоматическими повторами.
// Имя метода сохранено для совместимости с SenderService.
func (s *GRPCSender) SendJSONWithRetries(metrics ...*models.Metrics) error {
	if len(metrics) == 0 {
		return nil
	}

	req := buildUpdateMetricsRequest(metrics)
	return s.sendWithRetries(req)
}

func (s *GRPCSender) sendWithRetries(req *pb.UpdateMetricsRequest) error {
	const (
		maxAttempts = 4
		deltaDelay  = 2 * time.Second
	)
	currDelay := 1 * time.Second
	var lastErr error

	for attempt := 0; attempt < maxAttempts; attempt++ {
		if attempt != 0 {
			time.Sleep(currDelay)
			currDelay += deltaDelay
		}

		err := s.doSingleCall(req)
		if err == nil {
			return nil
		}

		if !isRetriableGRPCError(err) {
			return err
		}
		lastErr = err
		logger.Log.Info("повтор gRPC-запроса", logger.Err(err))
	}
	return fmt.Errorf("не удалось отправить метрики по gRPC после %d попыток: %w", maxAttempts, lastErr)
}

func (s *GRPCSender) doSingleCall(req *pb.UpdateMetricsRequest) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if s.localIP != "" {
		ctx = metadata.AppendToOutgoingContext(ctx, realIPMetadataKey, s.localIP)
	}

	_, err := s.client.UpdateMetrics(ctx, req)
	return err
}

// Close закрывает gRPC-соединение.
func (s *GRPCSender) Close() error {
	if s.conn == nil {
		return nil
	}
	return s.conn.Close()
}

func buildUpdateMetricsRequest(metrics []*models.Metrics) *pb.UpdateMetricsRequest {
	out := make([]*pb.Metric, 0, len(metrics))
	for _, m := range metrics {
		out = append(out, modelToProto(m))
	}
	return &pb.UpdateMetricsRequest{Metrics: out}
}

func modelToProto(m *models.Metrics) *pb.Metric {
	pm := &pb.Metric{Id: m.ID}
	switch m.MType {
	case models.Gauge:
		pm.Type = pb.Metric_GAUGE
		if m.Value != nil {
			pm.Value = *m.Value
		}
	case models.Counter:
		pm.Type = pb.Metric_COUNTER
		if m.Delta != nil {
			pm.Delta = *m.Delta
		}
	}
	return pm
}

// isRetriableGRPCError определяет, имеет ли смысл повторять gRPC-запрос.
// Повторяем сетевые ошибки и временные сбои сервиса (Unavailable, DeadlineExceeded).
func isRetriableGRPCError(err error) bool {
	if err == nil {
		return false
	}
	st, ok := status.FromError(err)
	if !ok {
		// не gRPC-ошибка, скорее всего сетевая
		return !errors.Is(err, context.Canceled)
	}
	switch st.Code() {
	case codes.Unavailable, codes.DeadlineExceeded, codes.ResourceExhausted, codes.Aborted:
		return true
	default:
		return false
	}
}
