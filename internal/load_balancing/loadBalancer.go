package loadbalancing

import (
	"errors"
	"net/url"
	"sync"
)

type LoadBalancer struct {
	Servers []*url.URL
	index   int
	mu      sync.Mutex
}

func NewLoadBalancer(servers []*url.URL) *LoadBalancer {
	return &LoadBalancer{Servers: servers, index: -1}
}

func (lb *LoadBalancer) GetNextServer() (*url.URL, error) {
	lb.mu.Lock()
	defer lb.mu.Unlock()

	if len(lb.Servers) == 0 {
		return nil, errors.New("no origin servers configured")
	}

	lb.index = (lb.index + 1) % len(lb.Servers)
	if lb.Servers[lb.index] == nil {
		return nil, errors.New("origin server is nil")
	}
	return lb.Servers[lb.index], nil
}
