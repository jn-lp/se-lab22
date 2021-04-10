package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"time"
)

var (
	timeout     = time.Duration(*timeoutSec) * time.Second
	serversPool = []string{
		"server1:8080",
		"server2:8080",
		"server3:8080",
	}

	reqCount = 0
)

func scheme() string {
	if *https {
		return "https"
	}
	return "http"
}

func health(dst string) (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(
		ctx,
		"GET",
		fmt.Sprintf("%s://%s/health", scheme(), dst),
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

func forward(dst string, rw http.ResponseWriter, r *http.Request) error {
	reqCount++

	ctx, cancel := context.WithTimeout(r.Context(), timeout)
	defer cancel()

	fwdRequest := r.Clone(ctx)
	fwdRequest.RequestURI = ""
	fwdRequest.URL.Host = dst
	fwdRequest.URL.Scheme = scheme()
	fwdRequest.Host = dst

	fwdRequest.Header.Set("lb-author", "rapid")
	fwdRequest.Header.Set("lb-req-cnt", strconv.Itoa(reqCount))

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
		rw.Header().Set("lb-from", dst)
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
