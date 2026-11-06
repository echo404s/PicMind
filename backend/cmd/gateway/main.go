package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/echo404s/PicMind/backend/gateway/config"
	"github.com/echo404s/PicMind/backend/gateway/router"
	"github.com/gin-gonic/gin"
)

func main() {
	env := "dev"
	if v := os.Getenv("APP_ENV"); v != "" {
		env = v
	}
	cfgPath := filepath.Join("config", "gateway", env+".yaml")

	cfg, err := config.LoadConfig(cfgPath)
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	gin.SetMode(cfg.Server.Mode)
	r := gin.Default()
	router.Register(r.Group("/gateway"))

	srv := &http.Server{
		Addr:              fmt.Sprintf(":%d", cfg.Server.Port),
		Handler:           r,
		ReadHeaderTimeout: 5 * time.Second,
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	go func() {
		log.Printf("%s listening on :%d (config=%s)", cfg.Service.Name, cfg.Server.Port, cfgPath)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("serve: %v", err)
		}
	}()

	<-ctx.Done()
	log.Println("shutting down...")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = srv.Shutdown(shutdownCtx)
}
