package registry

import "fmt"

// Discover returns healthy (passing) instances for serviceName.
func (c *Client) Discover(serviceName string) ([]Instance, error) {
	if serviceName == "" {
		return nil, fmt.Errorf("discover: serviceName is required")
	}
	entries, _, err := c.api.Health().Service(serviceName, "", true, nil)
	if err != nil {
		return nil, fmt.Errorf("consul discover %s: %w", serviceName, err)
	}
	out := make([]Instance, 0, len(entries))
	for _, e := range entries {
		out = append(out, Instance{
			ID:      e.Service.ID,
			Address: e.Service.Address,
			Port:    e.Service.Port,
			Tags:    e.Service.Tags,
		})
	}
	return out, nil
}

// Pick returns one healthy instance using round-robin.
func (c *Client) Pick(serviceName string) (Instance, error) {
	instances, err := c.Discover(serviceName)
	if err != nil {
		return Instance{}, err
	}
	if len(instances) == 0 {
		return Instance{}, fmt.Errorf("no healthy instances for %s", serviceName)
	}
	idx := c.rr.Add(1) - 1
	return instances[idx%uint64(len(instances))], nil
}
