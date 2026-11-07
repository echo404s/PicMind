package registry

import "fmt"

// Check configures a Consul health check for a registered service.
// Set exactly one of HTTP, GRPC, or TCP.
type Check struct {
	HTTP                           string
	GRPC                           string
	TCP                            string
	Interval                       string
	Timeout                        string
	DeregisterCriticalServiceAfter string
}

const (
	defaultInterval = "10s"
	defaultTimeout  = "5s"
	defaultDereg    = "1m"
)

// HTTPCheck probes url on Interval (for Gin / HTTP services).
func HTTPCheck(url string) Check {
	return Check{
		HTTP:                           url,
		Interval:                       defaultInterval,
		Timeout:                        defaultTimeout,
		DeregisterCriticalServiceAfter: defaultDereg,
	}
}

// GRPCCheck probes addr (host:port) via gRPC Health (for gRPC services).
func GRPCCheck(addr string) Check {
	return Check{
		GRPC:                           addr,
		Interval:                       defaultInterval,
		Timeout:                        defaultTimeout,
		DeregisterCriticalServiceAfter: defaultDereg,
	}
}

// TCPCheck probes host:port connectivity only.
func TCPCheck(addr string) Check {
	return Check{
		TCP:                            addr,
		Interval:                       defaultInterval,
		Timeout:                        defaultTimeout,
		DeregisterCriticalServiceAfter: defaultDereg,
	}
}

func (c Check) validate() error {
	n := 0
	if c.HTTP != "" {
		n++
	}
	if c.GRPC != "" {
		n++
	}
	if c.TCP != "" {
		n++
	}
	if n != 1 {
		return fmt.Errorf("check: set exactly one of HTTP, GRPC, TCP")
	}
	if c.Interval == "" || c.Timeout == "" {
		return fmt.Errorf("check: Interval and Timeout are required")
	}
	return nil
}
