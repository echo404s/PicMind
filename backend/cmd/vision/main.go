package main

import (
	"context"
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

	"github.com/echo404s/PicMind/backend/pkg/registry"
	"github.com/echo404s/PicMind/backend/vision/config"
	"github.com/echo404s/PicMind/backend/vision/server"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
)

func main() {
	env := "dev"
	if v := os.Getenv("APP_ENV"); v != "" {
		env = v
	}
	cfgPath := filepath.Join("config", "vision", env+".yaml")

	cfg, err := config.LoadConfig(cfgPath)
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	// gRPC
	grpcLis, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.Server.GRPCPort))
	if err != nil {
		log.Fatalf("grpc listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	hs := health.NewServer()
	healthpb.RegisterHealthServer(grpcServer, hs)
	hs.SetServingStatus("", healthpb.HealthCheckResponse_SERVING)
	hs.SetServingStatus(cfg.Service.Name, healthpb.HealthCheckResponse_SERVING)
	// Business RPCs register here later; reuse httpServer.Services().

	// --- HTTP (api-gateway) ---
	httpApp := server.New(cfg)
	httpServer := &http.Server{
		Addr:              fmt.Sprintf(":%d", cfg.Server.HTTPPort),
		Handler:           httpApp.Handler(),
		ReadHeaderTimeout: 5 * time.Second,
	}

	reg, err := registry.New(cfg.Consul.Address, cfg.Consul.Token)
	if err != nil {
		log.Fatalf("registry: %v", err)
	}

	grpcID := cfg.Service.ID + "-grpc"
	httpID := cfg.Service.ID + "-http"
	httpName := cfg.Service.Name + "-http"
	grpcName := cfg.Service.Name + "-grpc"

	grpcCheckAddr := fmt.Sprintf("%s:%d", cfg.Service.Address, cfg.Server.GRPCPort)
	if err := reg.Register(registry.Service{
		ID:      grpcID,
		Name:    grpcName,
		Address: cfg.Service.Address,
		Port:    cfg.Server.GRPCPort,
		Tags:    appendTags(cfg.Service.Tags, "grpc"),
	}, registry.GRPCCheck(grpcCheckAddr)); err != nil {
		log.Fatalf("register grpc: %v", err)
	}

	httpCheckURL := fmt.Sprintf("http://%s:%d/healthz", cfg.Service.Address, cfg.Server.HTTPPort)
	if err := reg.Register(registry.Service{
		ID:      httpID,
		Name:    httpName,
		Address: cfg.Service.Address,
		Port:    cfg.Server.HTTPPort,
		Tags:    appendTags(cfg.Service.Tags, "http"),
	}, registry.HTTPCheck(httpCheckURL)); err != nil {
		_ = reg.Deregister(grpcID)
		log.Fatalf("register http: %v", err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	errCh := make(chan error, 2)
	go func() {
		log.Printf("%s gRPC listening on :%d (consul=%s)", cfg.Service.Name, cfg.Server.GRPCPort, grpcName)
		errCh <- grpcServer.Serve(grpcLis)
	}()
	go func() {
		log.Printf("%s HTTP listening on :%d (consul=%s)", cfg.Service.Name, cfg.Server.HTTPPort, httpName)
		if err := httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
	}()

	select {
	case <-ctx.Done():
		log.Println("shutting down...")
	case err := <-errCh:
		log.Printf("serve stopped: %v", err)
	}

	hs.SetServingStatus("", healthpb.HealthCheckResponse_NOT_SERVING)
	hs.SetServingStatus(cfg.Service.Name, healthpb.HealthCheckResponse_NOT_SERVING)

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = httpServer.Shutdown(shutdownCtx)
	grpcServer.GracefulStop()

	if err := reg.Deregister(grpcID); err != nil {
		log.Printf("deregister grpc: %v", err)
	}
	if err := reg.Deregister(httpID); err != nil {
		log.Printf("deregister http: %v", err)
	}
}

func appendTags(base []string, extra string) []string {
	out := make([]string, 0, len(base)+1)
	out = append(out, base...)
	out = append(out, extra)
	return out
}
