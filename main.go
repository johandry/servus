package main

import (
	"fmt"
	"net"
	"os"
	"regexp"
	"strconv"

	"github.com/hashicorp/serf/serf"
	"github.com/johandry/servus/version"
)

const (
	defaultBindAddr = "0.0.0.0"
	defaultBindPort = 9000
)

var validAddr = regexp.MustCompile(`^([^:]*)(:[0-9]*)?$`)

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
	bindAddr, err := getBindAddr()
	if err != nil {
		return err
	}

	nodeName, err := os.Hostname()
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
	conf.MemberlistConfig.BindAddr = bindAddr.IP.String()
	conf.MemberlistConfig.BindPort = bindAddr.Port
	conf.NodeName = nodeName

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

func getBindAddr() (*net.TCPAddr, error) {
	envBindAddr := os.Getenv("SERVUS_BIND_ADDRESS")
	reAddr := validAddr.FindStringSubmatch(envBindAddr)
	if len(reAddr) != 3 {
		// SERVUS_BIND_ADDRESS is not a valid address
		envBindAddr = defaultBindAddr + ":" + strconv.Itoa(defaultBindPort)
	} else {
		if reAddr[1] == "" {
			reAddr[1] = defaultBindAddr
		}
		if reAddr[2] == "" || reAddr[2] == ":" {
			reAddr[2] = ":" + strconv.Itoa(defaultBindPort)
		}
		envBindAddr = reAddr[1] + reAddr[2]
	}

	bindIP, strPort, _ := net.SplitHostPort(envBindAddr)
	bindPort, _ := strconv.Atoi(strPort)
	if len(bindIP) != 0 && bindIP != defaultBindAddr && net.ParseIP(bindIP) != nil {
		return &net.TCPAddr{
			IP:   net.ParseIP(bindIP),
			Port: bindPort,
		}, nil
	}
	// return nil, fmt.Errorf("Cannot identify the bind address from 'SERVUS_BIND_ADDRESS' (%s)", envBindAddr)

	// Get interfaces to identify a valid IP address
	ifaces, err := net.Interfaces()
	if err != nil {
		return nil, err
	}
	// Check every interface to find a valid IP address
	for _, i := range ifaces {
		addrs, neterr := i.Addrs()
		if neterr != nil {
			continue
		}
		// Check every IP in that interface
		for _, a := range addrs {
			if ipnet, ok := a.(*net.IPNet); ok {
				// Cannot be a loopback IP, have to be a IPv4 address, and cannot be a
				// link-local unicast address
				if !ipnet.IP.IsLoopback() && ipnet.IP.To4() != nil && !ipnet.IP.IsLinkLocalUnicast() {
					// In case it is needed: iface = i.Name
					// If this is a valid IP, stop the search
					return &net.TCPAddr{
						IP:   ipnet.IP,
						Port: bindPort,
					}, nil
				}
			}
		}
	}

	// Bind Address not found
	return nil, fmt.Errorf("Cannot identify the bind address from 'SERVUS_BIND_ADDRESS' (%s) neither a network interface", envBindAddr)
}

func main() {
	version.Println()

	// (Command.Run)
	// 1. Read/parse config (Command.readConfig) -> config
	// 		1.1. Read parameters from every way
	// 		1.2. Get default config (Agent.DefaultConfig)
	// 		1.3. Set nodeName `os.Hostname()`
	// 		1.4. Set event handlers
	// 2. Setup loggers (Command.setupLoggers(config)) -> logger
	// 3. Setup agent (Command.setupAgent(config, logger)) -> agent
	// 		3.1. Get bindAddr IP:Port
	// 		3.2. Get advertiseAddr IP:Port
	//    3.3. Get encrypt keys
	// 		3.4. Get Serf config (Serf.DefaultConfig) -> serfConfig
	// 		3.5. Set serfConfig.MemberlistConfig default config (memberlist.DefaultLANConfig, memberlist.DefaultWANConfig, memberlist.DefaultLocalConfig)
	// 		3.6. Set serfConfig.MemberlistConfig. BindAddr, BindPort, AdvertiseAddr, AdvertisePort, SecretKey
	// 		3.7. serfConfig.NodeName = config.NodeName
	// 		3.8. serfConfig.Tags = config.Tags
	//		3.9. Create agent (Agent.Create(config, serfConfig, logger)) -> agent
	// 				3.9.1. Create channel for Serf events: `eventCh := make(chan serf.Event, 64); serfConfig.EventCh = eventCh`
	// 				3.9.2. Setup agent `agent := &Agent{ ... }`
	// 				3.9.3. Read & set tags from file *REAPEATED in 3.8?*
	// 				3.9.4. Read & set encrypt keys from file *REAPEATED in 3.3?*
	// 4. Defer Shutdown (agent.Shutdown)
	// 		4.1. Lock
	// 		4.2. Defer unlock
	// 		4.3. Serf.Shutdown
	// 		4.4. Close agent shutdown channel `Agent.shutdownCh`
	// 5. Start agent (Command.startAgent(agent, config, logger)) -> ipc
	// 		5.1. Register the event handler (Agent.RegisterEventHandler(&ScriptEventHandler{ .. }))
	// 		5.2. Start agent (Agent.Start())
	// 				5.2.1. Create serf (Serf.Create(Agent.config)) -> serf
	// 				5.2.2. Save serf to agent `Agent.serf = serf`
	// 				5.2.3. Go event loop (go Agent.ventLoop())
	// 						5.2.3.1. Endless loop to get info from the event channel and
	// 											agent and serf shutdown channels.
	// 						5.2.3.2. When a event is received: Lock, get the event handlers [from 1.4],
	// 											unlock and handle the event (ScriptEventHandler.HandleEvent(event))
	// 								5.2.3.2.1. if it's a valid event (invoke())
	// 								5.2.3.2.2. execute script (invokeEventScript(logger, script, self, event) @ invoke.go)
	// 						5.2.3.3. When a serf shutdown is received: shutdown agent
	// 											(Agent.Shutdown()) and exit loop/function
	// 						5.2.3.4. When an agent shutdown is received: exit loop/function
	// 		5.3. Get bindAddr IP:Port for logging
	// 		5.4. Setup RPC listener
	// 		5.5. Start IPC/RPC (NewAgentIPC(agent, key, listener, logger)) -> ipc
	// 				5.5.1. Create &AgentIPC{ ... }
	// 				5.5.2. Go IPC listen (AgentIPC.listen()) *No go dipper bc I'll implement RPC or something else*
	// 		5.6. Get advertiseAddr IP:Port for logging
	// 		5.6. Log status, return ipc
	// 6. Defer IPC shutdown (ipc.Shutdown())
	// 		6.1. Close IPC channel
	// 		6.2. Close listener
	// 		6.1. Close every client connection
	// 7. Join (Command.startupJoin(agent, config))
	// 		7.1. Agent join (Agent.Join(addresses)) -> n
	// 				7.1.1. Serf join (Serf.Join(addresses)) -> n
	// 8. Go retry joins (Command.retryJoin(agent, config, channel))
	// 		8.1. Endless loop to join agent (Agent.Join(addresses) *Refer 6.1.1*)
	// 		8.2. If succeed, return. Else, try again until max attempts
	// 9. Handle signals (Command.handleSignals(agent, config, channel))
	// 		9.1. Wait for a OS signal `os/signal.Notify()`
	// 		9.2. When Agent shutdown, the signal is os.Interrupt
	// 		9.3. When retry join fail, return
	// 		9.3. When Agent shutdown, return
	// 		9.3. If signal is syscall.SIGHUP, reload configuration (Command.handleReload(config, agent)) -> config
	// 				9.3.1. Read config file (Command.readConfig *Refer 1*)
	// 				9.3.2. Update event handlers
	// 				9.3.2. Update tags
	// 		9.4. If leave gracefully [631 @ command.go], leave agent (Agent.Leave())
	// 				9.4.1. Serf leave (Agent.Serf.Leave())
	// 		9.5. Wait for a second signal or timeout to return and exit agent.

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
