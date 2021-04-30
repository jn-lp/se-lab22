package integration

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	. "gopkg.in/check.v1"
)

const baseAddress = "http://balancer:8090"

var client = http.Client{
	Timeout: 3 * time.Second,
}

func Test(t *testing.T) { TestingT(t) }

type IntegrationSuite struct{}

var _ = Suite(&IntegrationSuite{})

func (s *IntegrationSuite) TestBalancer(c *C) {
	var responser string
	for i := 0; i < 9; i++ {
		resp, err := client.Get(fmt.Sprintf("%s/api/v1/some-data", baseAddress))
		c.Assert(err, IsNil)

		c.Assert(resp.StatusCode, Equals, http.StatusOK)

		if from := resp.Header.Get("lb-from"); responser == "" {
			responser = from
		} else {
			c.Assert(responser, Equals, from)
		}
	}
}

func (s *IntegrationSuite) BenchmarkBalancer(c *C) {
	for i := 0; i < c.N; i++ {
		resp, err := client.Get(fmt.Sprintf("%s/api/v1/some-data", baseAddress))
		c.Assert(err, IsNil)
		c.Assert(resp.StatusCode, Equals, http.StatusOK)
	}
}
