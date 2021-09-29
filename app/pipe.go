// +build windows

package app

import (
	"bufio"
	"log"

	"github.com/mysteriumnetwork/myst-launcher/model"

	"github.com/Microsoft/go-winio"
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

func StopApp() bool {
	pipe, err := winio.DialPipe(LauncherPipeName, nil)
	if err == nil {
		pipe.Write([]byte("stop\n"))
		return true
	}
	return false
}

func CreatePipeAndListen(ui model.Gui_) {
	l, err := winio.ListenPipe(LauncherPipeName, nil)
	if err != nil {
		log.Fatal(err)
	}

	handleCommand := func ()  {
		c, err := l.Accept()
		if err != nil {
			panic(err)
		}
		defer c.Close()

		rw := bufio.NewReadWriter(bufio.NewReader(c), bufio.NewWriter(c))
		s, _ := rw.ReadString('\n')
		switch s {
		case "popup\n":
			ui.ShowMain()

		case "stop\n":
			ui.TerminateWaitDialogueComplete()
			ui.CloseUI()
		}	
	}

	go func() {
		for {
			handleCommand()
		}
	}()
}
