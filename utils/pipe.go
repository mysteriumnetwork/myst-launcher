package utils

import (
	"bufio"
	"log"

	"github.com/Microsoft/go-winio"
	"github.com/mysteriumnetwork/myst-launcher/gui"
)

var LauncherPipeName = `\\.\pipe\mysterium_node_launcher`

func IsAlreadyRunning() bool {
	pipe, err := winio.DialPipe(LauncherPipeName, nil)
	if err == nil {
		pipe.Write([]byte("popup\n"))
		return true
	}
	return false
}

func CreatePipeAndListen(model *gui.UIModel) {
	l, err := winio.ListenPipe(LauncherPipeName, nil)
	if err != nil {
		log.Fatal(err)
	}

	go func() {
		for {
			c, err := l.Accept()
			if err != nil {
				panic(err)
			}
			defer c.Close()

			rw := bufio.NewReadWriter(bufio.NewReader(c), bufio.NewWriter(c))
			s, _ := rw.ReadString('\n')
			if s == "popup\n" {
				model.ShowMain()
			}
		}
	}()
}
