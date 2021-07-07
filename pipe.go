package main

import (
	"bufio"
	"log"
	"net"

	"github.com/Microsoft/go-winio"
)

var testPipeName = `\\.\pipe\mysterium_node_launcher`

func initPipe() bool {
	pipe, err := winio.DialPipe(testPipeName, nil)
	if err == nil {
		pipe.Write([]byte("popup\n"))
		return false
	} else {
		mod.pipeListener, err = winio.ListenPipe(testPipeName, nil)
		if err != nil {
			log.Fatal(err)
		}
		go server(mod.pipeListener)
	}
	return true
}

func server(l net.Listener) {
	for {
		c, err := l.Accept()
		if err != nil {
			panic(err)
		}
		rw := bufio.NewReadWriter(bufio.NewReader(c), bufio.NewWriter(c))
		s, _ := rw.ReadString('\n')
		if s == "popup\n" {
			mod.ShowMain()
		}
		c.Close()
	}
}
