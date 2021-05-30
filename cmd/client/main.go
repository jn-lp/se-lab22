package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"time"
)

func main() {
	target := flag.String(
		"target",
		"http://localhost:8090",
		"request target",
	)

	flag.Parse()

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	for range time.Tick(time.Second) {
		resp, err := client.Get(fmt.Sprintf("%s/api/v1/some-data", *target))
		if err != nil {
			log.Printf("error %s", err)

			continue
		}

		err = resp.Body.Close()
		if err != nil {
			log.Printf("error %s", err)

			continue
		}

		log.Printf("response %d", resp.StatusCode)
	}
}
