package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/echo404s/PicMind/backend/pkg/registry"
	"github.com/echo404s/PicMind/backend/vision"
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

	cfg, err := vision.LoadConfig(cfgPath)
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.Server.Port))
	if err != nil {
		log.Fatalf("listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	hs := health.NewServer()
	healthpb.RegisterHealthServer(grpcServer, hs)
	hs.SetServingStatus("", healthpb.HealthCheckResponse_SERVING)
	hs.SetServingStatus(cfg.Service.Name, healthpb.HealthCheckResponse_SERVING)
	// Business RPCs (Score / Describe) register here after proto is ready.

	reg, err := registry.New(cfg.Consul.Address, cfg.Consul.Token)
	if err != nil {
		log.Fatalf("registry: %v", err)
	}

	grpcCheckAddr := fmt.Sprintf("%s:%d", cfg.Service.Address, cfg.Server.Port)
	if err := reg.Register(registry.Service{
		ID:      cfg.Service.ID,
		Name:    cfg.Service.Name,
		Address: cfg.Service.Address,
		Port:    cfg.Server.Port,
		Tags:    cfg.Service.Tags,
	}, registry.GRPCCheck(grpcCheckAddr)); err != nil {
		log.Fatalf("register: %v", err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	errCh := make(chan error, 1)
	go func() {
		log.Printf("%s gRPC listening on :%d (config=%s)", cfg.Service.Name, cfg.Server.Port, cfgPath)
		errCh <- grpcServer.Serve(lis)
	}()

	select {
	case <-ctx.Done():
		log.Println("shutting down...")
	case err := <-errCh:
		log.Printf("serve stopped: %v", err)
	}

	hs.SetServingStatus("", healthpb.HealthCheckResponse_NOT_SERVING)
	hs.SetServingStatus(cfg.Service.Name, healthpb.HealthCheckResponse_NOT_SERVING)
	if err := reg.Deregister(cfg.Service.ID); err != nil {
		log.Printf("deregister: %v", err)
	}
	grpcServer.GracefulStop()
}
