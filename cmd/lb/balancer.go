package main

import (
	"context"
	"crypto/sha1"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

const (
	author = "rapid"
)

// ErrNoServersAreAlive for servers that are not alive from last pick.
var ErrNoServersAreAlive = errors.New("no servers are alive")

type LoadBalancer struct {
	pool    []*Server
	timeout time.Duration

	reqCount int
}

func NewLoadBalancer(timeout time.Duration) *LoadBalancer {
	return &LoadBalancer{
		timeout: timeout,
	}
}

func (l *LoadBalancer) SetServers(urls ...string) {
	for _, rawURL := range urls {
		u, _ := url.Parse(rawURL)
		l.pool = append(l.pool, New(u))
	}
}

func (l *LoadBalancer) Start(healthEvery time.Duration) {
	for i, srv := range l.pool {
		serverToCheck := srv

		j := i

		go func() {
			for range time.Tick(healthEvery) {
				alive, err := l.health(serverToCheck.String())
				if err != nil {
					continue
				}

				l.pool[j].SetAlive(alive)
			}
		}()
	}
}

func (l *LoadBalancer) Proxy(rw http.ResponseWriter, r *http.Request) error {
	ctx, cancel := context.WithTimeout(r.Context(), l.timeout)
	defer cancel()

	dst, err := l.pick(r.URL)
	if err != nil {
		return err
	}

	fwdRequest := r.Clone(ctx)
	fwdRequest.RequestURI = ""
	fwdRequest.URL.Host = dst.URL.String()
	fwdRequest.URL.Scheme = scheme()
	fwdRequest.Host = dst.URL.String()

	fwdRequest.Header.Set("lb-author", author)
	fwdRequest.Header.Set("lb-req-cnt", strconv.Itoa(l.reqCount))

	resp, err := http.DefaultClient.Do(fwdRequest)
	if err != nil {
		log.Printf("Failed to get response from %s: %s", dst, err)
		rw.WriteHeader(http.StatusServiceUnavailable)

		return err
	}

	for k, values := range resp.Header {
		for _, value := range values {
			rw.Header().Add(k, value)
		}
	}

	if *traceEnabled {
		rw.Header().Set("lb-from", dst.URL.Host)
	}

	log.Println("fwd", resp.StatusCode, resp.Request.URL)
	rw.WriteHeader(resp.StatusCode)

	_ = resp.Body.Close()

	_, err = io.Copy(rw, resp.Body)
	if err != nil {
		log.Printf("Failed to write response: %s", err)
	}

	return nil
}

func (l *LoadBalancer) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	if err := l.Proxy(rw, r); err != nil {
		rw.WriteHeader(http.StatusServiceUnavailable)
		log.Printf("Got error proxying request: %v", err)
	}

	l.reqCount++
}

func (l *LoadBalancer) health(dst string) (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), l.timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(
		ctx,
		"GET",
		fmt.Sprintf("%s://%s/health", scheme(), dst),
		nil,
	)
	if err != nil {
		return false, err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return false, err
	}

	_ = resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return false, nil
	}

	return true, nil
}

func (l *LoadBalancer) pick(url *url.URL) (*Server, error) {
	poolLen := len(l.pool)
	aliveCount := 0

	for _, srv := range l.pool {
		if srv.Alive() {
			aliveCount++
		}
	}

	if aliveCount == 0 {
		return nil, ErrNoServersAreAlive
	}

	// This loop may require too many cycles if low percentage of servers is alive
	// If only 1% is alive around 50% of request will require more than 70 cycles
	// With 3 servers and worst case scenario of 33.(3)% it won't be that bad
	// And solving this cutting edge case is beyond the scope of a task
	for nonce := 0; true; nonce++ {
		index := hash(url.RequestURI()+strconv.Itoa(nonce)) % uint64(poolLen)
		if srv := l.pool[index]; srv.Alive() {
			return srv, nil
		}
	}

	return nil, ErrNoServersAreAlive
}

func hash(s string) uint64 {
	h := sha1.Sum([]byte(s))

	return binary.BigEndian.Uint64(h[:8])
}
