package main

import (
	"context"
	"fmt"
	"hash/fnv"
	"io"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/jn-lp/se-lab22/server"
)

var author = "rapid"

type LoadBalancer struct {
	pool    []*server.Server
	timeout time.Duration

	reqCount int
}

func NewLoadBalancer(timeout time.Duration) *LoadBalancer {
	return &LoadBalancer{
		timeout: timeout,
	}
}

func (l LoadBalancer) SetServers(urls ...string) {
	for _, rawUrl := range urls {
		u, _ := url.Parse(rawUrl)
		l.pool = append(l.pool, server.New(u))
	}
}

func (l LoadBalancer) Start(healthEvery time.Duration) error {
	for i, srv := range l.pool {
		serverToCheck := srv
		i := i
		go func() {
			for range time.Tick(healthEvery) {
				alive, err := l.health(serverToCheck.String())
				if err != nil {
					continue
				}
				l.pool[i].SetAlive(alive)
			}
		}()
	}

	return nil
}

func (l LoadBalancer) Proxy(rw http.ResponseWriter, r *http.Request) error {
	ctx, cancel := context.WithTimeout(r.Context(), l.timeout)
	defer cancel()

	dst := l.pick(r.URL).URL

	fwdRequest := r.Clone(ctx)
	fwdRequest.URL = dst
	fwdRequest.Host = dst.Host

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
		rw.Header().Set("lb-from", dst.Host)
	}
	log.Println("fwd", resp.StatusCode, resp.Request.URL)
	rw.WriteHeader(resp.StatusCode)

	defer func(Body io.ReadCloser) {
		err = Body.Close()
	}(resp.Body)

	_, err = io.Copy(rw, resp.Body)
	if err != nil {
		log.Printf("Failed to write response: %s", err)
	}

	return nil
}

func (l LoadBalancer) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	if err := l.Proxy(rw, r); err != nil {
		log.Printf("Got error proxying request: %v", err)
	}
}

func (l LoadBalancer) health(dst string) (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), l.timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(
		ctx,
		"GET",
		fmt.Sprintf("%s/health", dst),
		nil,
	)
	if err != nil {
		return false, nil
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return false, nil
	}

	if resp.StatusCode != http.StatusOK {
		return false, nil
	}

	return true, nil
}

func (l LoadBalancer) pick(url *url.URL) *server.Server {
	return l.pool[hash(url.Path)%uint32(len(l.pool))]
}

func hash(s string) uint32 {
	h := fnv.New32a()
	_, err := h.Write([]byte(s))
	if err != nil {
		return 0
	}
	return h.Sum32()
}
