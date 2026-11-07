package registry

import (
	"fmt"
	"sync/atomic"

	"github.com/hashicorp/consul/api"
)

// Service describes an instance advertised to Consul.
type Service struct {
	ID      string
	Name    string
	Address string
	Port    int
	Tags    []string
}

// Instance is a healthy service endpoint from discovery.
type Instance struct {
	ID      string
	Address string
	Port    int
	Tags    []string
}

// Addr returns host:port for dialing.
func (i Instance) Addr() string {
	return fmt.Sprintf("%s:%d", i.Address, i.Port)
}

// Registrar registers and deregisters a service with Consul.
type Registrar interface {
	Register(svc Service, check Check) error
	Deregister(serviceID string) error
}

// Discovery looks up healthy instances of a named service.
type Discovery interface {
	Discover(serviceName string) ([]Instance, error)
	Pick(serviceName string) (Instance, error)
}

// Client wraps a Consul API client for register + discover.
type Client struct {
	api *api.Client
	rr  atomic.Uint64
}

// New creates a Consul-backed registry client.
func New(addr, token string) (*Client, error) {
	if addr == "" {
		return nil, fmt.Errorf("consul address is required")
	}
	c := api.DefaultConfig()
	c.Address = addr
	if token != "" {
		c.Token = token
	}
	cli, err := api.NewClient(c)
	if err != nil {
		return nil, fmt.Errorf("consul client: %w", err)
	}
	return &Client{api: cli}, nil
}

var (
	_ Registrar = (*Client)(nil)
	_ Discovery = (*Client)(nil)
)
