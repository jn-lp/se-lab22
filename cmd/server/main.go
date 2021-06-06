package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/jn-lp/se-lab22/cmd"
	"github.com/jn-lp/se-lab22/httptools"
	"github.com/jn-lp/se-lab22/signal"
)

const (
	teamName  = "rapid"
	dbAddress = "http://db:8070"
	// confResponseDelaySec = "CONF_RESPONSE_DELAY_SEC"
	confHealthFailure = "CONF_HEALTH_FAILURE"
)

func main() {
	port := flag.Int(
		"port",
		8080,
		"server port",
	)

	flag.Parse()
	putTeam()

	h := http.NewServeMux()

	h.HandleFunc(
		"/health",
		func(rw http.ResponseWriter, r *http.Request) {
			rw.Header().Set("content-type", "text/plain")
			if failConfig := os.Getenv(confHealthFailure); failConfig == "true" {
				rw.WriteHeader(http.StatusInternalServerError)
				_, _ = rw.Write([]byte("FAILURE"))
			} else {
				rw.WriteHeader(http.StatusOK)
				_, _ = rw.Write([]byte("OK"))
			}
		},
	)

	report := make(Report)

	h.HandleFunc(
		"/api/v1/some-data",
		func(rw http.ResponseWriter, r *http.Request) {
			key := r.FormValue("key")
			if key == "" {
				rw.WriteHeader(http.StatusNotFound)

				return
			}

			resp, err := http.Get(fmt.Sprintf("%s/db/%s", dbAddress, key))
			if err != nil {
				rw.WriteHeader(http.StatusInternalServerError)

				return
			}

			for k, values := range resp.Header {
				for _, value := range values {
					rw.Header().Add(k, value)
				}
			}

			rw.WriteHeader(resp.StatusCode)
			defer resp.Body.Close()

			if _, err = io.Copy(rw, resp.Body); err != nil {
				return
			}
		},
	)

	h.Handle("/report", report)

	server := httptools.CreateServer(*port, h)
	server.Start()

	signal.WaitForTerminationSignal()
}

func putTeam() {
	req := cmd.PutRequest{
		Value: []byte(time.Now().Format("2021-04-25")),
	}

	reqJSON, err := json.Marshal(req)
	if err != nil {
		log.Fatal(err)
	}

	res, err := http.Post(
		fmt.Sprintf("%s/db/%s", dbAddress, teamName),
		"application/json",
		bytes.NewBuffer(reqJSON),
	)
	if err != nil {
		log.Fatal(err)
	}

	if res.StatusCode != http.StatusOK {
		log.Fatalf("cannot put timestamp: %s\n", res.Status)
	}
}
