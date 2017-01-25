package test

import (
	"eaciit/sebard/modules"
	"testing"
)

const (
	host1 = "localhost:8888"
	host2 = "localhost:8889"

	host1config = "../config/node0.json"
	host2config = "../config/node1.json"
)

var (
	s1, s2 *modules.SebarNode
)

func TestStartHost(t *testing.T) {
	s1 = new(modules.SebarNode)
	s1.ReadConfig(host1config)
	s1.Start()
}

func TestStartClient(t *testing.T) {
	s2 = new(modules.SebarNode)
	s2.ReadConfig(host2config)
	s2.Start()
}

func TestCloseHost(t *testing.T) {
	if s2 != nil {
		s2.Close()
	}
	s1.Close()
}
