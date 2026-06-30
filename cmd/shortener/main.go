package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/TMWF/url-shortener/internal/audit"
	"github.com/TMWF/url-shortener/internal/config"
	"github.com/TMWF/url-shortener/internal/database"
	"github.com/TMWF/url-shortener/internal/grpcserver"
	"github.com/TMWF/url-shortener/internal/grpcserver/interceptors"
	"github.com/TMWF/url-shortener/internal/handler"
	"github.com/TMWF/url-shortener/internal/logger"
	"github.com/TMWF/url-shortener/internal/middleware"
	"github.com/TMWF/url-shortener/internal/proto"
	"github.com/TMWF/url-shortener/internal/repository"
	"github.com/TMWF/url-shortener/internal/service"
	"github.com/TMWF/url-shortener/internal/util"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

var (
	buildVersion string = "N/A"
	buildDate    string = "N/A"
	buildCommit  string = "N/A"
)

func main() {
	printBuildInfo()
	cfg := config.InitialiseConfigs()

	err := logger.Initialize(cfg.LogLevel)
	if err != nil {
		logger.GetLogger().Fatal("Error occured while initialising logger")
	}
	defer logger.GetLogger().Sync()

	db, err := database.GetDB(cfg.DatabaseDSN)
	if err != nil && !errors.Is(err, database.ErrEmptyDSN) {
		logger.GetLogger().Fatal("Error occured when creating DB connection")
	}
	if db != nil {
		defer func() {
			if err := db.Close(); err != nil {
				logger.GetLogger().Error(
					"Failed to properly close the db connection",
					zap.String("original error message", err.Error()),
				)
			}
			logger.GetLogger().Debug("Successfully closed sql/db")
		}()
	}

	var auditFile *os.File

	if cfg.AuditFileStoragePath != "" {
		file, err := os.OpenFile(cfg.AuditFileStoragePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			log.Fatal(err)
		}
		defer file.Close()
	}

	router, grpcServer := createRouterAndGrpcServer(cfg, db, auditFile)
	logger.GetLogger().Info("Starting server on port " + cfg.ServerHost)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	defer stop()

	ctx, cancel := context.WithCancelCause(ctx)
	defer cancel(nil)

	server := &http.Server{
		Addr:    cfg.ServerHost,
		Handler: router,
	}

	grpcListener, err := net.Listen("tcp", cfg.GrpcAddress)
	if err != nil {
		logger.GetLogger().Fatal("Failed to listen for gRPC", zap.Error(err))
	}

	go func() {
		logger.GetLogger().Info("Starting gRPC server", zap.String("address", cfg.GrpcAddress))
		if err := grpcServer.Serve(grpcListener); err != nil && !errors.Is(err, grpc.ErrServerStopped) {
			cancel(err)
		}
	}()

	go func() {
		if cfg.EnableHttps {
			util.GenerateCertificate(cancel, cfg)

			homeDir, err := os.UserHomeDir()
			if err != nil {
				cancel(fmt.Errorf("http server error: %w", err))
			}

			err = server.ListenAndServeTLS(
				filepath.Join(homeDir, cfg.CertFilepath),
				filepath.Join(homeDir, cfg.KeyFilePath),
			)

			if err != nil && !errors.Is(err, http.ErrServerClosed) {
				cancel(fmt.Errorf("http server error: %w", err))
			}

		} else {
			err = server.ListenAndServe()

			if err != nil && !errors.Is(err, http.ErrServerClosed) {
				cancel(fmt.Errorf("http server error: %w", err))
			}
		}
	}()

	<-ctx.Done()
	logger.GetLogger().Info("Shutting down gracefully...", zap.String("reason", ctx.Err().Error()))

	shutdownCtx, shutdownRelease := context.WithTimeout(context.Background(), 15*time.Second)
	defer shutdownRelease()

	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.GetLogger().Error("HTTP server shutdown error", zap.Error(err))
	} else {
		logger.GetLogger().Info("HTTP server stopped successfully")
	}

	logger.GetLogger().Info("Stopping gRPC server...")
	grpcServer.GracefulStop()
}

func createRouterAndGrpcServer(config *config.Config, db *sql.DB, auditFile *os.File) (http.Handler, *grpc.Server) {
	storage := repository.GetStorage(config, db)
	jwtHelper := util.NewJWTHelper(config)
	urlService := service.NewURLService(storage, config)
	urlHandler := handler.NewURLHandlerWithTrustedSubnet(urlService, jwtHelper, config.TrustedSubnet)

	if auditFile != nil {
		localAuditeventHandler := audit.NewLocalAuditEventHandler(auditFile)
		urlHandler.RegisterObserver(localAuditeventHandler)
	}

	if config.AuditURL != "" {
		remoteAuditEventHandler := audit.NewRemoteAuditEventHandler(config.AuditURL)
		urlHandler.RegisterObserver(remoteAuditEventHandler)
	}

	grpcServer := grpc.NewServer(grpc.UnaryInterceptor(interceptors.AuthInterceptor(config)))
	proto.RegisterShortenerServiceServer(grpcServer, grpcserver.NewShortenerGRPCServer(urlService))

	router := chi.NewRouter()
	router.Use(middleware.JwtTokenMiddleware(config))
	router.Use(middleware.RequestLoggerMiddleware(logger.GetLogger()))
	router.Use(middleware.GzipMiddleware())
	router.Post(`/`, urlHandler.ShortenURL)
	router.Post(`/api/shorten`, urlHandler.ShortenURLAPI)
	router.Post(`/api/shorten/batch`, urlHandler.ShortenURLBatch)
	router.Get(`/api/user/urls`, urlHandler.GetUserURLs)
	router.Delete(`/api/user/urls`, urlHandler.DeleteUserURLs)
	router.Get(`/{id}`, urlHandler.GetOriginalURL)
	router.Get(`/api/internal/stats`, urlHandler.GetStats)

	if dbStorage, ok := storage.(repository.DBStorage); ok {
		pingService := service.NewPingDBService(dbStorage)
		pingHandler := handler.NewPingHandler(pingService)
		router.Get(`/ping`, pingHandler.PingDB)
	}

	return router, grpcServer
}

func printBuildInfo() {
	fmt.Printf("Build version: %s\n", buildVersion)
	fmt.Printf("Build date: %s\n", buildDate)
	fmt.Printf("Build commit: %s\n", buildCommit)
}
