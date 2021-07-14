package utils

import (
	"bufio"
	"errors"
	"github.com/Microsoft/go-winio"
	"github.com/mysteriumnetwork/myst-launcher/gui"
	"golang.org/x/sys/windows"
	"log"
)

var LauncherPipeName = `\\.\pipe\mysterium_node_launcher`

func IsAlreadyRunning() bool {
	_, err := winio.DialPipe(LauncherPipeName, nil)
	return !errors.Is(err, windows.ERROR_FILE_NOT_FOUND)
}

func CreatePipeAndListen(model *gui.UIModel) {
	l, err := winio.ListenPipe(LauncherPipeName, nil)
	if err != nil {
		log.Fatal(err)
	}

	go func() {
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
	}()
}
