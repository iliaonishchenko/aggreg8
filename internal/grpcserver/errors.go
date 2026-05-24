package grpcserver

import "errors"

var (
	errEmptyMetric = errors.New("метрика отсутствует")
	errMissingID   = errors.New("отсутствует идентификатор метрики")
	errUnknownType = errors.New("неизвестный тип метрики")
)
