package main

import (
	"bytes"
	"fmt"
	. "gopkg.in/check.v1"
	"net/http"
	"testing"
)

func Test(t *testing.T) { TestingT(t) }

type MySuite struct{}

var _ = Suite(&MySuite{})

func (s *MySuite) TestServerPicker(c *C) {
	request, _ := http.NewRequest(
		http.MethodGet,
		"https://example.com:8080/uri",
		bytes.NewReader([]byte("body")))
	c.Assert(pickServer(request, []string{}, []bool{}), Equals, "")
	c.Assert(pickServer(request, []string{"server"}, []bool{false}), Equals, "")
	c.Assert(pickServer(request, []string{"server"}, []bool{true}), Equals, "server")
	c.Assert(pickServer(request, []string{"first", "second"}, []bool{false, true}), Equals, "second")
	delta := 0
	for i := 0; i < 10000; i++ {
		request, _ := http.NewRequest(
			http.MethodGet,
			fmt.Sprintf("https://example.com:8080/uri%d", i),
			bytes.NewReader([]byte("body")))
		server := pickServer(request, []string{"first", "second", "third"}, []bool{true, false, true})
		c.Assert(server != "second", Equals, true)
		if server == "first" {
			delta++
		}
		if server == "third" {
			delta--
		}
	}
	// According to http://probability.ca/jeff/writing/margerror.html it will work 19 times out of 20
	// For the single hash function it doesn't matter cause the result will always be the same
	// But if hash function is refactored there is 5% chance that it will fail for no reason
	c.Assert(delta > -100 || delta < 100, Equals, true)
}
