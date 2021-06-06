package main

import (
	"flag"
	"log"
	"time"

	"github.com/jn-lp/se-lab22/httptools"
	"github.com/jn-lp/se-lab22/signal"
)

var serversPool = []string{
	"server1:8080",
	"server2:8080",
	"server3:8080",
}

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

func scheme() string {
	if *https {
		return "https"
	}

	return "http"
}

func main() {
	flag.Parse()

	lb := NewLoadBalancer(time.Duration(*timeoutSec) * time.Second)
	lb.SetServers(serversPool...)

	lb.Start(10 * time.Second)

	frontend := httptools.CreateServer(*port, lb)

	log.Println("Starting load balancer...")
	log.Printf("Tracing support enabled: %t", *traceEnabled)
	frontend.Start()

	signal.WaitForTerminationSignal()
}
