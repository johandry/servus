package main

import (
	"log"
	"os"
	"time"

	"github.com/hashicorp/serf/serf"
	"github.com/johandry/servus/version"
)

// Servus ...
type Servus struct {
	conf       *serf.Config
	serf       *serf.Serf
	leaderAddr string
}

// CreateServus ...
func CreateServus() (servus *Servus, err error) {
	servus = &Servus{}

	leaderAddr := os.Getenv("SERVUS_LEADER_ADDRESS")
	advertiseAddr := os.Getenv("SERVUS_ADVERTISE_ADDRESS")
	bindAddr := os.Getenv("SERVUS_BIND_ADDRESS")
	rpcAddr := os.Getenv("SERVUS_RPC_ADDRESS")

	if leaderAddr == "" {
		leaderAddr = "127.0.0.1"
	}
	if advertiseAddr == "" {
		advertiseAddr = "127.0.0.1"
	}
	if bindAddr == "" {
		bindAddr = "0.0.0.0"
	}
	if rpcAddr == "" {
		rpcAddr = "0.0.0.0"
	}

	servus.leaderAddr = leaderAddr

	servus.conf = serf.DefaultConfig()
	servus.conf.MemberlistConfig.BindAddr = bindAddr
	servus.conf.MemberlistConfig.AdvertiseAddr = advertiseAddr
	servus.conf.MemberlistConfig.ProbeInterval = 50 * time.Millisecond
	servus.conf.MemberlistConfig.ProbeTimeout = 25 * time.Millisecond
	servus.conf.MemberlistConfig.SuspicionMult = 1
	servus.conf.NodeName = servus.conf.MemberlistConfig.BindAddr
	servus.conf.Tags = map[string]string{"role": "servus", "servus": "servus", "tag01": "foo"}

	servus.serf, err = serf.Create(servus.conf)
	if err != nil {
		return servus, err
	}
	return servus, nil
}

// Start ...
func (servus *Servus) Start() {
	_, err := servus.serf.Join([]string{servus.leaderAddr}, true)
	if err != nil {
		log.Printf("Could not join to the servus at %s, creating a new one. %v\n", servus.leaderAddr, err)
	}
	defer servus.serf.Leave()

	ch := make(chan string)
	for {
		select {
		case msg := <-ch:
			log.Printf("Message received: %s\n", msg)
		case <-time.After(time.Second * 60):
			log.Println("Timeout")
			servus.serf.Leave()
			return
		}
	}
}

func main() {
	version.Println()
	servus, err := CreateServus()
	if err != nil {
		log.Fatalln(err)
	}
	servus.Start()
}
