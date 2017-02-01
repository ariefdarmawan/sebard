package test

import (
	"eaciit/sebard/modules"
	"testing"
	"time"

	"github.com/eaciit/toolkit"
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
	idx := 0
	s1.AddEvent("healthcheck", func(in toolkit.M) *toolkit.Result {
		sn := in.Get("server").(*modules.SebarNode)
		idx++
		sn.Log().Info(toolkit.Sprintf("Health check %s idx: %d", sn.Config.HostAddress(), idx))
		return nil
	})
	s1.Start()
}

func TestCloseHostAfterWait(t *testing.T) {
	go func() {
		s1.Wait()
	}()

	time.Sleep(10 * time.Second)
	s1.SendCloseSignal()
}

func TestStartClient(t *testing.T) {
	s1 = new(modules.SebarNode)
	s1.ReadConfig(host1config)
	s1.AddEvent("healthcheck", func(in toolkit.M) *toolkit.Result {
		sn := in.Get("server").(*modules.SebarNode)
		sn.Log().Info(toolkit.Sprintf("Health check %s", sn.Config.HostAddress()))
		return nil
	})
	s1.Start()

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
