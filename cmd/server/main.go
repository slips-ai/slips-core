package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	taskv1 "github.com/slips-ai/slips-core/gen/api/proto/task/v1"
	tagv1 "github.com/slips-ai/slips-core/gen/api/proto/tag/v1"

	taskapp "github.com/slips-ai/slips-core/internal/task/application"
	taskgrpc "github.com/slips-ai/slips-core/internal/task/infra/grpc"
	taskpg "github.com/slips-ai/slips-core/internal/task/infra/postgres"

	tagapp "github.com/slips-ai/slips-core/internal/tag/application"
	taggrpc "github.com/slips-ai/slips-core/internal/tag/infra/grpc"
	tagpg "github.com/slips-ai/slips-core/internal/tag/infra/postgres"

	"github.com/slips-ai/slips-core/pkg/auth"
	"github.com/slips-ai/slips-core/pkg/config"
	"github.com/slips-ai/slips-core/pkg/logger"
	"github.com/slips-ai/slips-core/pkg/tracing"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

// chainUnaryInterceptors chains multiple unary interceptors into one
func chainUnaryInterceptors(interceptors ...grpc.UnaryServerInterceptor) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		chain := handler
		for i := len(interceptors) - 1; i >= 0; i-- {
			interceptor := interceptors[i]
			next := chain
			chain = func(currentCtx context.Context, currentReq interface{}) (interface{}, error) {
				return interceptor(currentCtx, currentReq, info, next)
			}
		}
		return chain(ctx, req)
	}
}

func main() {
	// Load configuration
	cfg, err := config.Load("")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Initialize logger
	isDev := os.Getenv("ENV") != "production"
	logr := logger.New(isDev)
	slog.SetDefault(logr)

	logr.Info("Starting slips-core service", "port", cfg.Server.GRPCPort)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Initialize tracing
	var shutdown func(context.Context) error
	if cfg.Tracing.Enabled {
		shutdown, err = tracing.InitTracer(cfg.Tracing.ServiceName, cfg.Tracing.Endpoint)
		if err != nil {
			logr.Warn("Failed to initialize tracing", "error", err)
		} else {
			defer func() {
				// Use a fresh context with timeout for tracer shutdown
				shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer shutdownCancel()
				if err := shutdown(shutdownCtx); err != nil {
					logr.Error("Failed to shutdown tracer", "error", err)
				}
			}()
			logr.Info("Tracing initialized", "endpoint", cfg.Tracing.Endpoint)
		}
	}

	// Connect to database
	dbpool, err := pgxpool.New(ctx, cfg.Database.DatabaseURL())
	if err != nil {
		logr.Error("Failed to connect to database", "host", cfg.Database.Host, "error", err)
		os.Exit(1)
	}
	defer dbpool.Close()

	if err := dbpool.Ping(ctx); err != nil {
		logr.Error("Failed to ping database", "host", cfg.Database.Host, "error", err)
		os.Exit(1)
	}
	logr.Info("Database connected", "host", cfg.Database.Host)

	// Initialize JWT validator
	jwtValidator := auth.NewJWTValidator(cfg.Auth.JWKSEndpoint, cfg.Auth.ExpectedIssuer)
	
	// Fetch JWKS keys
	// NOTE: Keys are only fetched at startup. In production, implement periodic refresh
	// or on-demand fetching when unknown 'kid' is encountered to handle key rotation.
	if err := jwtValidator.FetchJWKS(ctx); err != nil {
		logr.Error("Failed to fetch JWKS", "error", err)
		os.Exit(1)
	}
	logr.Info("JWT validator initialized", "jwks_endpoint", cfg.Auth.JWKSEndpoint)

	// Initialize repositories
	taskRepo := taskpg.NewTaskRepository(dbpool)
	tagRepo := tagpg.NewTagRepository(dbpool)

	// Initialize services
	taskService := taskapp.NewService(taskRepo, logr)
	tagService := tagapp.NewService(tagRepo, logr)

	// Initialize gRPC servers
	taskServer := taskgrpc.NewTaskServer(taskService)
	tagServer := taggrpc.NewTagServer(tagService)

	// Create gRPC server with interceptors
	var opts []grpc.ServerOption
	
	// Add interceptors in order: auth first, then tracing
	// Auth runs first to reject unauthenticated requests before creating trace spans
	opts = append(opts, grpc.UnaryInterceptor(chainUnaryInterceptors(
		auth.UnaryServerInterceptor(jwtValidator),
		tracing.UnaryServerInterceptor(),
	)))
	
	grpcServer := grpc.NewServer(opts...)

	// Register services
	taskv1.RegisterTaskServiceServer(grpcServer, taskServer)
	tagv1.RegisterTagServiceServer(grpcServer, tagServer)

	// Register reflection service for grpcurl and other tools
	reflection.Register(grpcServer)

	// Start gRPC server
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.Server.GRPCPort))
	if err != nil {
		logr.Error("Failed to listen", "error", err)
		os.Exit(1)
	}

	// Handle graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigChan
		logr.Info("Shutting down gracefully...")
		grpcServer.GracefulStop()
		cancel()
	}()

	logr.Info("gRPC server listening", "address", lis.Addr())
	if err := grpcServer.Serve(lis); err != nil {
		logr.Error("Failed to serve", "error", err)
		os.Exit(1)
	}
}
