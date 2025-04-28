package interceptors

import (
	"ResuMatch/internal/utils"
	"context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

func RequestIDClientInterceptor() grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		if requestID := utils.GetRequestID(ctx); requestID != "" {
			ctx = metadata.AppendToOutgoingContext(ctx, "X-Request-ID", requestID)
		}
		return invoker(ctx, method, req, reply, cc, opts...)
	}
}

func RequestIDServerInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		if md, ok := metadata.FromIncomingContext(ctx); ok {
			if values := md.Get("X-Request-ID"); len(values) > 0 {
				ctx = utils.SetRequestID(ctx, values[0])
			}
		}
		return handler(ctx, req)
	}
}
