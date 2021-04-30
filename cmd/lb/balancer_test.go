package main

import (
	"net/url"
	"testing"
	"time"

	"github.com/jn-lp/se-lab22/server"
	. "gopkg.in/check.v1"
)

func Test(t *testing.T) { TestingT(t) }

type TestSuite struct{}

var _ = Suite(&TestSuite{})

func (s *TestSuite) TestBalancer(c *C) {
	for _, test := range []struct {
		URL         *url.URL
		ServersPool []*server.Server
		ServerIndex int
		WithError   bool
	}{} {
		lb := NewLoadBalancer(time.Duration(*timeoutSec) * time.Second)
		lb.pool = test.ServersPool

		srv, err := lb.pick(test.URL)
		c.Assert(err != nil, Equals, test.WithError)
		c.Assert(srv, Equals, test.ServersPool[test.ServerIndex])
	}
}
