package grpcserver

import (
	"context"
	"errors"

	"github.com/TMWF/url-shortener/internal/proto"
	"github.com/TMWF/url-shortener/internal/repository"
	"github.com/TMWF/url-shortener/internal/service"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

type ShortenerGRPCServer struct {
	proto.UnimplementedShortenerServiceServer
	urlService service.URLService
}

func NewShortenerGRPCServer(service service.URLService) *ShortenerGRPCServer {
	return &ShortenerGRPCServer{
		urlService: service,
	}
}

func (s *ShortenerGRPCServer) ShortenURL(ctx context.Context, req *proto.URLShortenRequest) (*proto.URLShortenResponse, error) {
	if req.GetUrl() == "" {
		return nil, status.Error(codes.InvalidArgument, "url is required")
	}

	shortURL, err := s.urlService.ShortenURL(ctx, req.GetUrl())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to shorten URL: %v", err)
	}

	return proto.URLShortenResponse_builder{Result: shortURL}.Build(), nil
}

func (s *ShortenerGRPCServer) ExpandURL(ctx context.Context, req *proto.URLExpandRequest) (*proto.URLExpandResponse, error) {
	if req.GetId() == "" {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}

	originalURL, err := s.urlService.GetOriginalURL(ctx, req.GetId())
	if err != nil {
		if errors.Is(err, repository.ErrURLNotFound) {
			return nil, status.Error(codes.NotFound, "URL not found")
		}

		return nil, status.Errorf(codes.Internal, "failed to expand URL: %v", err)
	}

	return proto.URLExpandResponse_builder{Result: originalURL}.Build(), nil
}

func (s *ShortenerGRPCServer) ListUserURLs(ctx context.Context, _ *emptypb.Empty) (*proto.UserURLsResponse, error) {
	userURLs, err := s.urlService.GetUserURLs(ctx)
	if err != nil {
		if errors.Is(err, repository.ErrUserIDAbsent) {
			return nil, status.Error(codes.Unauthenticated, "user unauthorized")
		}
		return nil, status.Errorf(codes.Internal, "failed to fetch user URLs: %v", err)
	}

	response := proto.UserURLsResponse_builder{
		Url: make([]*proto.URLData, 0, len(userURLs)),
	}.Build()

	for _, item := range userURLs {
		response.SetUrl(
			append(
				response.GetUrl(),
				proto.URLData_builder{
					ShortUrl:    item.ShortURL,
					OriginalUrl: item.OriginalURL,
				}.Build(),
			),
		)
	}

	return response, nil
}
