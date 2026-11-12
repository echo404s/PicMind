package service

// Services aggregates use-case services for vision-service.
// Add new services as fields here instead of growing constructor parameter lists.
type Services struct {
	// Score *ScoreService
}

// NewServices wires all use-case services from shared dependencies.
func NewServices() *Services {
	return &Services{}
}
