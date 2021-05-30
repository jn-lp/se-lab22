package integration

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	. "gopkg.in/check.v1"
)

const baseAddress = "http://balancer:8090"

func Test(t *testing.T) {
	time.Sleep(10 * time.Second)

	TestingT(t)
}

type IntegrationSuite struct{}

var _ = Suite(&IntegrationSuite{})

func (s *IntegrationSuite) TestBalancer(c *C) {
	client := http.Client{
		Timeout: 3 * time.Second,
	}

	var srv string

	for i := 0; i < 9; i++ {
		resp, err := client.Get(fmt.Sprintf("%s/api/v1/some-data", baseAddress))
		if err != nil {
			continue
		}

		err = resp.Body.Close()
		if err != nil {
			continue
		}

		c.Assert(err, IsNil)
		c.Assert(resp.StatusCode, Equals, http.StatusOK)

		if from := resp.Header.Get("lb-from"); srv == "" {
			srv = from
		} else {
			c.Assert(srv, Equals, from)
		}
	}
}

func (s *IntegrationSuite) BenchmarkBalancer(c *C) {
	client := http.Client{
		Timeout: 3 * time.Second,
	}

	for i := 0; i < c.N; i++ {
		resp, err := client.Get(fmt.Sprintf("%s/api/v1/some-data", baseAddress))
		c.Assert(err, IsNil)

		c.Assert(resp.Body.Close(), IsNil)

		c.Assert(resp.StatusCode, Equals, http.StatusOK)
	}
}
