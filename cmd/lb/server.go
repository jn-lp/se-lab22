package main

import (
	"net/url"
	"sync"
)

type Server struct {
	sync.RWMutex

	URL   *url.URL
	alive bool
}

func New(URL *url.URL) *Server {
	return &Server{URL: URL}
}

func (s *Server) SetAlive(alive bool) {
	s.Lock()
	defer s.Unlock()

	s.alive = alive
}

func (s *Server) Alive() bool {
	s.Lock()
	defer s.Unlock()

	return s.alive
}

func (s *Server) String() string {
	return s.URL.String()
}
