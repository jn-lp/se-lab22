package main

import (
	"net/url"
	"testing"
	"time"

	. "gopkg.in/check.v1"
)

func Test(t *testing.T) { TestingT(t) }

type TestSuite struct{}

var _ = Suite(&TestSuite{})

func (s *TestSuite) TestBalancer(c *C) {
	URL, err := url.Parse("https://example.com:8080/uri")
	if err != nil {
		return
	}

	for _, test := range []struct {
		ServersPool []*Server
		ServerIndex int
		WithError   bool
	}{
		{
			WithError: true,
		},
		{
			ServersPool: []*Server{
				{
					alive: false,
				},
			},
			WithError: true,
		},
		{
			ServersPool: []*Server{
				{
					alive: true,
				},
			},
			ServerIndex: 0,
			WithError:   false,
		},
		{
			ServersPool: []*Server{
				{
					alive: false,
				},
				{
					alive: true,
				},
			},
			ServerIndex: 1,
			WithError:   false,
		},
	} {
		lb := NewLoadBalancer(time.Duration(*timeoutSec) * time.Second)
		lb.pool = test.ServersPool

		srv, err := lb.pick(URL)
		c.Assert(err != nil, Equals, test.WithError)
		if !test.WithError {
			c.Assert(srv, Equals, test.ServersPool[test.ServerIndex])
		}
	}
}
