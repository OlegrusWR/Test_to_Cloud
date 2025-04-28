package backend

import (
	"net/http"
	"sync"
	"time"

	
)

type Server struct{
	URL string
	mu sync.RWMutex
	connCount int
	alive bool
	healthCheck *HealthCheck
}

type HealthCheck struct{
	Interval time.Duration
	Timeout time.Duration
	Path string
	Client *http.Client
}

func NewServer(url string, hc *HealthCheck) *Server {
	s := &Server{
		URL: url,
		alive: true,
		healthCheck: &HealthCheck{
			Interval: hc.Interval,
			Timeout: hc.Timeout,
			Path: hc.Path,
			Client: &http.Client{
				Timeout: hc.Timeout,
			},
		},
	}
	go s.StartCheck()
	return s
}