package integration

import (
	"fmt"
	"github.com/stretchr/testify/assert"
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

func (s *IntegrationSuite) TestBalancer(t *testing.T) {
	var serverName string
	for i := 0; i < 10; i++ {
		url := fmt.Sprintf("%s/api/v1/some-data?key=%s", baseAddress, ourTeam)

		resp, err := client.Get(url)
		assert.Nil(t, err)
		assert.Equal(t, resp.StatusCode, http.StatusOK)

		body, err := ioutil.ReadAll(resp.Body)
		assert.Nil(t, err)
		assert.NotEmpty(t, body)

		data := string(body)
		t.Log(fmt.Sprintf("body: %s", data))

		if i == 0 {
			serverName = resp.Header.Get("lb-from")
		} else {
			assert.Equal(t, serverName, resp.Header.Get("lb-from"))
		}
	}
	t.Log(fmt.Sprintf("server name: %s", serverName))
}

func (s *IntegrationSuite) BenchmarkBalancer(b *testing.B) {
	for i := 0; i < b.N; i++ {
		resp, err := client.Get(fmt.Sprintf("%s/api/v1/some-data?key=%s", baseAddress, ourTeam))
		assert.Nil(b, err)
		assert.Equal(b, resp.StatusCode, http.StatusOK)
	}
}

