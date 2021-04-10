package main

import (
	"encoding/json"
	"flag"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/jn-lp/se-lab22/httptools"
	"github.com/jn-lp/se-lab22/signal"
)

const (
	confResponseDelaySec = "CONF_RESPONSE_DELAY_SEC"
	confHealthFailure    = "CONF_HEALTH_FAILURE"
)

var port = flag.Int(
	"port",
	8080,
	"server port",
)

func main() {
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
			respDelayString := os.Getenv(confResponseDelaySec)
			if delaySec, parseErr := strconv.Atoi(respDelayString);
				parseErr == nil && delaySec > 0 && delaySec < 300 {
				time.Sleep(time.Duration(delaySec) * time.Second)
			}

			report.Process(r)

			rw.Header().Set("content-type", "application/json")
			rw.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(rw).Encode([]string{"1", "2"})
		},
	)

	h.Handle("/report", report)

	server := httptools.CreateServer(*port, h)
	server.Start()

	signal.WaitForTerminationSignal()
}
