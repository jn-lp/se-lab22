package main

import (
	"flag"
	"log"
	"net/http"
	"time"

	"github.com/jn-lp/se-lab22/httptools"
	"github.com/jn-lp/se-lab22/signal"
)

var (
	port = flag.Int(
		"port",
		8090,
		"load balancer port",
	)
	timeoutSec = flag.Int(
		"timeout-sec",
		3,
		"request timeout time in seconds",
	)
	https = flag.Bool(
		"https",
		false,
		"whether backends support HTTPs",
	)

	traceEnabled = flag.Bool(
		"trace",
		false,
		"whether to include tracing information into responses",
	)
)

func main() {
	flag.Parse()

	for i, server := range serversPool {
		i := i
		server := server
		go func() {
			for range time.Tick(10 * time.Second) {
				alive, err := health(server)
				if err != nil {
					return
				}
				serversMask[i] = alive
				log.Println(server, alive)
			}
		}()
	}

	frontend := httptools.CreateServer(
		*port,
		http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
			err := forward(pickServer(r, serversPool, serversMask), rw, r)
			if err != nil {
				rw.WriteHeader(http.StatusInternalServerError)

				return
			}
		}),
	)

	log.Println("Starting load balancer...")
	log.Printf("Tracing support enabled: %t", *traceEnabled)
	frontend.Start()

	signal.WaitForTerminationSignal()
}
