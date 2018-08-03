package main

import (
	"fmt"
	"os"

	"github.com/hashicorp/serf/serf"
)

var (
	seedPort = 7777
)

func main() {
	shutdownCh1 := make(chan struct{})
	shutdownCh2 := make(chan struct{})
	shutdownCh3 := make(chan struct{})

	go createNode("node1", seedPort, shutdownCh1)
	go createNode("node2", 7778, shutdownCh2)
	go createNode("node3", 7779, shutdownCh3)
	for {
		var ln string
		fmt.Scanln(&ln)
		switch ln {
		case "quit":
			os.Exit(0)
		case "kill1":
			close(shutdownCh1)
		case "start1":
			shutdownCh1 = make(chan struct{})
			go createNode("node1", seedPort, shutdownCh1)
		case "kill2":
			close(shutdownCh2)
		case "start2":
			shutdownCh2 = make(chan struct{})
			go createNode("node2", 7778, shutdownCh1)
		case "kill3":
			close(shutdownCh3)
		case "start3":
			shutdownCh3 = make(chan struct{})
			go createNode("node3", 7779, shutdownCh1)
		}
	}
}

func createNode(nodeName string, port int, shutdownCh chan struct{}) {
	ch := make(chan serf.Event, 256)

	config := serf.DefaultConfig()
	config.Init()
	config.MemberlistConfig.BindAddr = "127.0.0.1"
	config.MemberlistConfig.BindPort = port
	config.MemberlistConfig.AdvertiseAddr = "127.0.0.1"
	config.MemberlistConfig.AdvertisePort = port
	config.NodeName = nodeName
	config.EventCh = ch
	config.ProtocolVersion = 3

	s, err := serf.Create(config)
	if err != nil {
		panic(err)
	}

	_, err = s.Join([]string{fmt.Sprintf("127.0.0.1:%d", seedPort)}, true)
	if err != nil {
		fmt.Println("%s", err)
		return
	}

	for {
		select {
		case e := <-ch:
			switch e.EventType() {
			case serf.EventMemberJoin:
				fmt.Println("member count ", len(s.Members()))
			case serf.EventMemberLeave, serf.EventMemberFailed:
			case serf.EventMemberReap:
			}
		case _ = <-shutdownCh:
			//s.Leave()
			s.Shutdown()
			break
		}
	}
}
