package main

import (
	"fmt"
	"net"
	"os"
	"strconv"

	"github.com/hashicorp/serf/serf"
)

const (
	defaultBindAddr = "127.0.0.1"
	defaultBindPort = "9000"
)

// Serf encapsulate the Serf configuration
type Serf struct {
	serf        *serf.Serf
	addr        string
	eventCh     chan serf.Event
	initMembers []string
	shutdownCh  chan struct{}
}

// New create a Serf agent
func New() (*Serf, error) {
	serf := &Serf{
		shutdownCh: make(chan struct{}),
	}

	return serf, nil
}

// Start is to startup the Serf agent
func (s Serf) Start() error {
	eventCh := make(chan serf.Event, 64)

	// Get the bind address
	bindAddr := os.Getenv("SERVUS_BIND_ADDRESS")
	if len(bindAddr) == 0 {
		bindAddr = defaultBindAddr + ":" + defaultBindPort
	}
	addr, strPort, err := net.SplitHostPort(bindAddr)
	if err != nil {
		return err
	}
	if len(strPort) == 0 {
		strPort = defaultBindPort
	}
	port, err := strconv.Atoi(strPort)
	if err != nil {
		return err
	}

	conf := serf.DefaultConfig()
	// Get tags from config settigns
	conf.Tags = map[string]string{
		"role": "servus",
		"tag1": "foo",
		"tag2": "bar",
	}

	conf.Init()

	// Get these parameters from config settings
	conf.MemberlistConfig.BindAddr = addr
	conf.MemberlistConfig.BindPort = port
	conf.NodeName = bindAddr

	conf.EventCh = eventCh
	s.eventCh = eventCh

	srf, err := serf.Create(conf)
	if err != nil {
		return err
	}
	s.serf = srf

	// Join, move to func
	// Get members in the clusters to join
	leader := os.Getenv("SERVUS_LEADER_ADDRESS")
	initMembers := []string{leader}

	if len(leader) != 0 {
		s.initMembers = initMembers
		num, err := s.serf.Join(s.initMembers, true)
		if err != nil {
			return err
		}
		// Log?:
		fmt.Printf("Node join to the cluster with %d nodes", num)
	} else {
		// Log?:
		fmt.Print("First node in the cluster\n")
	}

	go s.serfEventHandlerLoop()

	return nil
}

func (s *Serf) serfEventHandlerLoop() {
	serfShutdownCh := s.serf.ShutdownCh()
	for {
		select {
		case e := <-s.eventCh:
			switch e.EventType() {
			case serf.EventMemberJoin:
				// Log and maybe more?:
				fmt.Printf("New member join %+v\n", e)
			case serf.EventMemberLeave:
				// Log and maybe more?:
				fmt.Printf("An existing member leave %+v\n", e)
			case serf.EventMemberFailed:
				// Log and maybe more?:
				fmt.Printf("A member have failed %+v\n", e)
			case serf.EventMemberUpdate:
				// Log and maybe more?:
				fmt.Printf("A member was updated %+v\n", e)
			case serf.EventMemberReap:
				// Log and maybe more?:
				fmt.Printf("A member was reaped %+v\n", e)
			case serf.EventUser:
				// Log and maybe more?:
				fmt.Printf("Received event: %+v\n", e)
			case serf.EventQuery:
				// Log and maybe more?:
				fmt.Printf("Received query: %+v\n", e)
			default:
				// Log and maybe more?:
				fmt.Printf("Received unknown serf event: %+v\n", e)
			}

		case <-serfShutdownCh:
			// Log and more todo:
			fmt.Print("Serf is shutting down, ... me too\n")
			close(s.shutdownCh)
			return

		case <-s.shutdownCh:
			return
		}
	}

}

func main() {
	s, err := New()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating Serf: %s\n", err)
	}
	err = s.Start()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error starting Serf: %s\n", err)
	}

	for {
	}
}
