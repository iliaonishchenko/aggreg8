// Package grpcserver реализует gRPC-сервис для приёма метрик от агентов.
package grpcserver

import (
	"context"
	"time"

	"github.com/iliaonishchenko/aggreg8/internal/audit"
	"github.com/iliaonishchenko/aggreg8/internal/logger"
	models "github.com/iliaonishchenko/aggreg8/internal/model"
	pb "github.com/iliaonishchenko/aggreg8/internal/proto"
	"github.com/iliaonishchenko/aggreg8/internal/service"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"
)

// MetricsServer реализует gRPC-сервис metrics.Metrics.
type MetricsServer struct {
	pb.UnimplementedMetricsServer
	storage  service.MetricStorage
	notifier audit.Notifier
}

// NewMetricsServer создаёт gRPC-сервис метрик, использующий переданное хранилище и нотификатор аудита.
func NewMetricsServer(storage service.MetricStorage, notifier audit.Notifier) *MetricsServer {
	return &MetricsServer{
		storage:  storage,
		notifier: notifier,
	}
}

// UpdateMetrics принимает батч метрик от агента, валидирует и сохраняет их в хранилище.
func (s *MetricsServer) UpdateMetrics(ctx context.Context, req *pb.UpdateMetricsRequest) (*pb.UpdateMetricsResponse, error) {
	if req == nil || len(req.GetMetrics()) == 0 {
		return nil, status.Error(codes.InvalidArgument, "пустой список метрик")
	}

	metrics := make([]*models.Metrics, 0, len(req.GetMetrics()))
	for _, m := range req.GetMetrics() {
		converted, err := protoToModel(m)
		if err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "некорректная метрика %q: %v", m.GetId(), err)
		}
		metrics = append(metrics, converted)
	}

	if err := s.storage.UpdateMetrics(metrics); err != nil {
		logger.Log.Error("ошибка сохранения батча метрик через gRPC", logger.Err(err))
		return nil, status.Error(codes.Internal, "не удалось сохранить метрики")
	}

	s.notifyAudit(ctx, metrics)

	return &pb.UpdateMetricsResponse{}, nil
}

func (s *MetricsServer) notifyAudit(ctx context.Context, metrics []*models.Metrics) {
	if s.notifier == nil {
		return
	}
	names := make([]string, 0, len(metrics))
	for _, m := range metrics {
		names = append(names, m.ID)
	}
	event := audit.AuditEvent{
		TS:        time.Now().Unix(),
		Metrics:   names,
		IPAddress: clientIP(ctx),
	}
	errs := s.notifier.NotifyAll(event)
	for _, err := range errs {
		logger.Log.Error("ошибка отправки аудит-события для gRPC батча", logger.Err(err))
	}
}

// clientIP извлекает IP клиента сначала из метаданных x-real-ip, затем из peer-адреса.
func clientIP(ctx context.Context) string {
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		if vals := md.Get("x-real-ip"); len(vals) > 0 && vals[0] != "" {
			return vals[0]
		}
	}
	if p, ok := peer.FromContext(ctx); ok && p.Addr != nil {
		return p.Addr.String()
	}
	return ""
}

// protoToModel преобразует protobuf-метрику во внутреннюю модель и валидирует обязательные поля.
func protoToModel(m *pb.Metric) (*models.Metrics, error) {
	if m == nil {
		return nil, errEmptyMetric
	}
	if m.GetId() == "" {
		return nil, errMissingID
	}
	switch m.GetType() {
	case pb.Metric_GAUGE:
		value := m.GetValue()
		return &models.Metrics{
			ID:    m.GetId(),
			MType: models.Gauge,
			Value: &value,
		}, nil
	case pb.Metric_COUNTER:
		delta := m.GetDelta()
		return &models.Metrics{
			ID:    m.GetId(),
			MType: models.Counter,
			Delta: &delta,
		}, nil
	default:
		return nil, errUnknownType
	}
}
