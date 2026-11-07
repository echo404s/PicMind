package registry

import (
	"fmt"

	"github.com/hashicorp/consul/api"
)

// Register advertises svc with the given health check.
func (c *Client) Register(svc Service, check Check) error {
	if svc.ID == "" {
		svc.ID = svc.Name
	}
	if svc.Name == "" {
		return fmt.Errorf("register: service Name is required")
	}
	if svc.Address == "" {
		return fmt.Errorf("register: service Address is required")
	}
	if svc.Port <= 0 {
		return fmt.Errorf("register: service Port must be > 0")
	}
	if err := check.validate(); err != nil {
		return err
	}

	agentCheck := &api.AgentServiceCheck{
		Interval:                       check.Interval,
		Timeout:                        check.Timeout,
		DeregisterCriticalServiceAfter: check.DeregisterCriticalServiceAfter,
	}
	switch {
	case check.HTTP != "":
		agentCheck.HTTP = check.HTTP
	case check.GRPC != "":
		agentCheck.GRPC = check.GRPC
		agentCheck.GRPCUseTLS = false
	case check.TCP != "":
		agentCheck.TCP = check.TCP
	}

	reg := &api.AgentServiceRegistration{
		ID:      svc.ID,
		Name:    svc.Name,
		Address: svc.Address,
		Port:    svc.Port,
		Tags:    svc.Tags,
		Check:   agentCheck,
	}
	if err := c.api.Agent().ServiceRegister(reg); err != nil {
		return fmt.Errorf("consul register %s: %w", svc.ID, err)
	}
	return nil
}

// Deregister removes the service instance from Consul.
func (c *Client) Deregister(serviceID string) error {
	if serviceID == "" {
		return fmt.Errorf("deregister: serviceID is required")
	}
	if err := c.api.Agent().ServiceDeregister(serviceID); err != nil {
		return fmt.Errorf("consul deregister %s: %w", serviceID, err)
	}
	return nil
}
