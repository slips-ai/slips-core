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
	authv1 "github.com/slips-ai/slips-core/gen/go/auth/v1"
	mcptokenv1 "github.com/slips-ai/slips-core/gen/go/mcptoken/v1"
	tagv1 "github.com/slips-ai/slips-core/gen/go/tag/v1"
	taskv1 "github.com/slips-ai/slips-core/gen/go/task/v1"

	mcptokenapp "github.com/slips-ai/slips-core/internal/mcptoken/application"
	mcptokengrpc "github.com/slips-ai/slips-core/internal/mcptoken/infra/grpc"
	mcptokenpg "github.com/slips-ai/slips-core/internal/mcptoken/infra/postgres"

	authapp "github.com/slips-ai/slips-core/internal/auth/application"
	authgrpc "github.com/slips-ai/slips-core/internal/auth/infra/grpc"
	authpg "github.com/slips-ai/slips-core/internal/auth/infra/postgres"

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

func main() {
	// Load configuration
	cfg, err := config.Load("config.yaml")
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

	// Initialize Identra gRPC client
	identraClient, err := auth.NewIdentraClient(cfg.Auth.IdentraGRPCEndpoint)
	if err != nil {
		logr.Error("Failed to initialize Identra client", "error", err)
		os.Exit(1)
	}
	defer identraClient.Close()
	logr.Info("Identra client initialized", "endpoint", cfg.Auth.IdentraGRPCEndpoint)

	// Initialize JWT validator
	jwtValidator := auth.NewJWTValidator(identraClient, cfg.Auth.ExpectedIssuer)

	// Fetch JWKS keys
	// NOTE: Keys are only fetched at startup. In production, implement periodic refresh
	// or on-demand fetching when unknown 'kid' is encountered to handle key rotation.
	if err := jwtValidator.FetchJWKS(ctx); err != nil {
		logr.Error("Failed to fetch JWKS", "error", err)
		os.Exit(1)
	}
	logr.Info("JWT validator initialized", "issuer", cfg.Auth.ExpectedIssuer)

	// Initialize repositories
	mcptokenRepo := mcptokenpg.NewMCPTokenRepository(dbpool)
	authRepo := authpg.NewRepository(dbpool)
	taskRepo := taskpg.NewTaskRepository(dbpool)
	tagRepo := tagpg.NewTagRepository(dbpool)

	// Initialize services
	mcptokenService := mcptokenapp.NewService(mcptokenRepo, logr)
	authService := authapp.NewService(
		authRepo,
		identraClient,
		cfg.Auth.OAuth.Provider,
		cfg.Auth.OAuth.RedirectURL,
		logr,
	)
	taskService := taskapp.NewService(taskRepo, tagRepo, logr)
	tagService := tagapp.NewService(tagRepo, logr)

	// Initialize gRPC servers
	mcptokenServer := mcptokengrpc.NewMCPTokenServer(mcptokenService)
	authServer := authgrpc.NewServer(authService)
	taskServer := taskgrpc.NewTaskServer(taskService)
	tagServer := taggrpc.NewTagServer(tagService)

	// Create gRPC server with interceptors
	var opts []grpc.ServerOption

	// Build interceptor chain in order: auth first, then (optionally) tracing
	// Auth runs first to reject unauthenticated requests before creating trace spans
	// Note: Auth interceptor automatically skips authentication for public Auth Service endpoints
	// (GetAuthorizationURL, HandleCallback, RefreshToken)
	interceptors := []grpc.UnaryServerInterceptor{
		auth.UnaryServerInterceptorWithMCP(jwtValidator, mcptokenService),
	}
	if cfg.Tracing.Enabled {
		interceptors = append(interceptors, tracing.UnaryServerInterceptor())
	}
	opts = append(opts, grpc.ChainUnaryInterceptor(interceptors...))
	grpcServer := grpc.NewServer(opts...)

	// Register services
	mcptokenv1.RegisterMCPTokenServiceServer(grpcServer, mcptokenServer)
	authv1.RegisterAuthServiceServer(grpcServer, authServer)
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
