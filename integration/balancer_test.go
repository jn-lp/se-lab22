package integration

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"testing"
	"time"

	. "gopkg.in/check.v1"
)

const baseAddress = "http://balancer:8090"
const ourTeam = "rapid"

var client = http.Client{
	Timeout: 3 * time.Second,
}

func Test(t *testing.T) {
	// Wait for servers to come up
	time.Sleep(10 * time.Second)

	TestingT(t)
}

type IntegrationSuite struct{}

var _ = Suite(&IntegrationSuite{})

func (s *IntegrationSuite) TestBalancer(c *C) {
	var serverName string
	for i := 0; i < 10; i++ {
		url := fmt.Sprintf("%s/api/v1/some-data?key=%s", baseAddress, ourTeam)

		resp, err := client.Get(url)
		c.Assert(err, IsNil)
		c.Assert(resp.StatusCode, Equals, http.StatusOK)

		body, err := ioutil.ReadAll(resp.Body)
		c.Assert(err, IsNil)

		data := string(body)
		c.Log(fmt.Sprintf("body: %s", data))

		if i == 0 {
			serverName = resp.Header.Get("lb-from")
		} else {
			c.Assert(serverName, Equals, resp.Header.Get("lb-from"))
		}
	}
	c.Log(fmt.Sprintf("server name: %s", serverName))
}

func (s *IntegrationSuite) BenchmarkBalancer(c *C) {
	for i := 0; i < c.N; i++ {
		resp, err := client.Get(fmt.Sprintf("%s/api/v1/some-data?key=%s", baseAddress, ourTeam))
		c.Assert(err, IsNil)
		c.Assert(resp.StatusCode, Equals, http.StatusOK)
	}
}

