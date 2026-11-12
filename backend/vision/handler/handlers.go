package handler

import "github.com/echo404s/PicMind/backend/vision/service"

// Handlers aggregates HTTP handlers for vision-service.
// Add new handlers as fields here instead of growing router.Register parameters.
type Handlers struct {
	// Score *ScoreHandler
}

// NewHandlers builds all HTTP handlers from services.
func NewHandlers(_ *service.Services) *Handlers {
	return &Handlers{}
}
