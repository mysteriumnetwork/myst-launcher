package main

import (
	"bufio"
	"log"
	"net"

	"github.com/microsoft/go-winio"
)

var testPipeName = `\\.\pipe\mysterium_node_launcher`

func initPipe() bool {
	pipe, err := winio.DialPipe(testPipeName, nil)
	//log.Printf("err> %v", err)
	//if err.(*os.PathError).Err != syscall.ENOENT {
	//	log.Fatalf("expected ENOENT got %v", err)
	//}
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

func server(l net.Listener) { // ch chan int
	for {
		c, err := l.Accept()
		if err != nil {
			panic(err)
		}
		rw := bufio.NewReadWriter(bufio.NewReader(c), bufio.NewWriter(c))
		s, err := rw.ReadString('\n')
		if err != nil {
			//fmt.Println(err)
			//panic(err)
		}

		if s == "popup\n" {
			//fmt.Println("popup")
			mod.ShowMain()
		}
		c.Close()
	}
}
