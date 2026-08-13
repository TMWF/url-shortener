package interceptors

import (
	"context"
	"strings"

	"github.com/TMWF/url-shortener/internal/config"
	"github.com/TMWF/url-shortener/internal/middleware"
	"github.com/TMWF/url-shortener/internal/util"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func AuthInterceptor(cfg *config.Config) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req any,
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (any, error) {
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return handler(ctx, req)
		}

		authHeader := md.Get("authorization")
		if len(authHeader) == 0 {
			return handler(ctx, req)
		}

		token := authHeader[0]
		token, _ = strings.CutPrefix(token, "Bearer ")

		userID, err := middleware.GetUserID(token, cfg)
		if err != nil {
			return nil, status.Error(codes.Unauthenticated, "invalid or expired token")
		}

		newCtx := context.WithValue(ctx, util.UserID, userID)

		return handler(newCtx, req)
	}
}
