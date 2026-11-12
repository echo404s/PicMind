package server

import (
	"net/http"

	"github.com/echo404s/PicMind/backend/vision/config"
	"github.com/echo404s/PicMind/backend/vision/handler"
	"github.com/echo404s/PicMind/backend/vision/router"
	"github.com/echo404s/PicMind/backend/vision/service"
	"github.com/gin-gonic/gin"
)

// Server owns HTTP wiring for vision-service: deps → services → handlers → routes.
type Server struct {
	engine   *gin.Engine
	services *service.Services
	handlers *handler.Handlers
}

// New builds the HTTP server from config (manual DI, no wire/fx).
func New(_ *config.Config) *Server {
	services := service.NewServices()
	handlers := handler.NewHandlers(services)

	gin.SetMode(gin.ReleaseMode)
	engine := gin.New()
	engine.Use(gin.Recovery())
	router.Register(engine, handlers)

	return &Server{
		engine:   engine,
		services: services,
		handlers: handlers,
	}
}

// Handler returns the root HTTP handler for ListenAndServe.
func (s *Server) Handler() http.Handler {
	return s.engine
}

// Services exposes use-case services (e.g. for future gRPC adapters).
func (s *Server) Services() *service.Services {
	return s.services
}
