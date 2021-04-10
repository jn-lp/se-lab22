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

	// TODO: Використовуйте дані про стан сервреа, щоб підтримувати список тих серверів, яким можна відправляти ззапит.
	for _, server := range serversPool {
		server := server
		go func() {
			for range time.Tick(10 * time.Second) {
				alive, err := health(server)
				if err != nil {
					return
				}
				log.Println(server, alive)
			}
		}()
	}

	frontend := httptools.CreateServer(
		*port,
		http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
			// TODO: Рееалізуйте свій алгоритм балансувальника.
			err := forward(serversPool[0], rw, r)
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
