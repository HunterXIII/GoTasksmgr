package interceptor

import (
	"context"
	"strings"
	"tasksmgr/contextx"

	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

func RequestIDInterceptor(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
	md, _ := metadata.FromIncomingContext(ctx)

	requestID := strings.Join(md.Get("x-request-id"), "")
	if requestID == "" {
		requestID = uuid.NewString()
	}
	// fmt.Printf("gRPC - RequstID: %s\n", requestID)
	ctx = context.WithValue(ctx, contextx.RequestIDKey{}, requestID)
	return handler(ctx, req)
}
