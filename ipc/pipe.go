//go:build windows
// +build windows

/**
 * Copyright (c) 2021 BlockDev AG
 *
 * This source code is licensed under the MIT license found in the
 * LICENSE file in the root directory of this source tree.
 */

package ipc

import (
	"bufio"
	"fmt"
	"net"

	"github.com/mysteriumnetwork/myst-launcher/model"

	"github.com/Microsoft/go-winio"
)

const launcherPipeName = `\\.\pipe\mysterium_node_launcher`

type Handler struct {
	pipe net.Listener
}

func NewHandler() *Handler {
	h := &Handler{}
	h.OpenPipe()
	return h
}

func (p *Handler) OwnsPipe() bool {
	return p.pipe != nil
}

func (p *Handler) OpenPipe() {
	l, _ := winio.ListenPipe(launcherPipeName, nil)
	p.pipe = l
}

func (p *Handler) Close() {
	if p.pipe != nil {
		p.pipe.Close()
	}
}

func (p *Handler) SendPopupApp() bool {
	pipe, err := winio.DialPipe(launcherPipeName, nil)
	if err == nil {
		pipe.Write([]byte("popup\n"))
		return true
	}
	return false
}

// send stop and own the pipe
func (p *Handler) SendStopApp() bool {
	pipe, err := winio.DialPipe(launcherPipeName, nil)
	if err == nil {
		pipe.Write([]byte("stop\n"))
		return true
	}
	return false
}

func (p *Handler) Listen(ui model.Gui_) {
	if p.pipe == nil {
		return
	}

	handleCommand := func() (exit bool) {
		c, err := p.pipe.Accept()
		if err != nil {
			fmt.Println("pipe > Listen !accept", err)
			exit = true
			return
		}
		defer c.Close()

		rw := bufio.NewReadWriter(bufio.NewReader(c), bufio.NewWriter(c))
		s, err := rw.ReadString('\n')
		if err != nil {
			fmt.Println("pipe > Listen !accept", err)
			exit = true
			return
		}
		fmt.Println("pipe >", s)

		switch s {
		case "popup\n":
			ui.PopupMain()

		case "stop\n":
			ui.CloseUI()
			exit = true
		}
		return
	}

	go func() {
		for {
			if handleCommand() == true {
				p.pipe.Close()
				break
			}
		}
	}()
}
