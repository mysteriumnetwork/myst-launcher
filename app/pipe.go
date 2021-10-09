//go:build windows
// +build windows

package app

import (
	"bufio"
	"net"

	"github.com/mysteriumnetwork/myst-launcher/model"

	"github.com/Microsoft/go-winio"
)

var LauncherPipeName = `\\.\pipe\mysterium_node_launcher`

type PipeHandler struct {
	pipe net.Listener
}

func NewPipeHandler() *PipeHandler {
	h := &PipeHandler{}
	h.OpenPipe()
	return h
}
func (p *PipeHandler) OwnsPipe() bool {
	return p.pipe != nil
}

func (p *PipeHandler) OpenPipe() {
	l, _ := winio.ListenPipe(LauncherPipeName, nil)
	p.pipe = l
}

func (p *PipeHandler) SendPopupApp() bool {
	pipe, err := winio.DialPipe(LauncherPipeName, nil)
	if err == nil {
		pipe.Write([]byte("popup\n"))
		return true
	}
	return false
}

// send stop and own the pipe
func  (p *PipeHandler) SendStopApp() bool {
	pipe, err := winio.DialPipe(LauncherPipeName, nil)
	if err == nil {
		pipe.Write([]byte("stop\n"))
		return true
	}
	return false
}

func (p *PipeHandler) Listen(ui model.Gui_) {
	if p.pipe == nil {
		return
	}

	handleCommand := func() {
		c, err := p.pipe.Accept()
		if err != nil {
			panic(err)
		}
		defer c.Close()

		rw := bufio.NewReadWriter(bufio.NewReader(c), bufio.NewWriter(c))
		s, _ := rw.ReadString('\n')
		switch s {
		case "popup\n":
			ui.PopupMain()

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
