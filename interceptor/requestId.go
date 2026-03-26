package interceptor

import (
	"context"
	"strings"

	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

type RequestIDKey struct{}

func RequestIDInterceptor(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
	md, _ := metadata.FromIncomingContext(ctx)

	requestID := strings.Join(md.Get("x-request-id"), "")
	if requestID == "" {
		requestID = uuid.NewString()
	}
	// fmt.Printf("gRPC - RequstID: %s\n", requestID)
	ctx = context.WithValue(ctx, RequestIDKey{}, requestID)
	return handler(ctx, req)
}
