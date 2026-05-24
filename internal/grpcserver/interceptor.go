package grpcserver

import (
	"context"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// RealIPMetadataKey — ключ метаданных gRPC, в котором агент передаёт свой IP-адрес.
const RealIPMetadataKey = "x-real-ip"

// TrustedSubnetInterceptor создаёт UnaryInterceptor, который проверяет,
// что IP агента (из метаданных x-real-ip) принадлежит указанной доверенной подсети.
// Если cidr пустой, возвращается nil — вызывающая сторона не должна регистрировать
// перехватчик, чтобы не добавлять лишний вызов на каждый RPC.
func TrustedSubnetInterceptor(cidr string) grpc.UnaryServerInterceptor {
	if cidr == "" {
		return nil
	}

	_, ipNet, err := net.ParseCIDR(cidr)
	if err != nil {
		return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
			return nil, status.Error(codes.PermissionDenied, "доверенная подсеть настроена некорректно")
		}
	}

	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, status.Error(codes.PermissionDenied, "метаданные отсутствуют")
		}
		vals := md.Get(RealIPMetadataKey)
		if len(vals) == 0 || vals[0] == "" {
			return nil, status.Error(codes.PermissionDenied, "метаданные x-real-ip обязательны")
		}
		ip := net.ParseIP(vals[0])
		if ip == nil {
			return nil, status.Error(codes.PermissionDenied, "метаданные x-real-ip некорректны")
		}
		if !ipNet.Contains(ip) {
			return nil, status.Error(codes.PermissionDenied, "IP агента вне доверенной подсети")
		}
		return handler(ctx, req)
	}
}
