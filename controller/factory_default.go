//go:build !linux
// +build !linux

package controller

import (
	"github.com/mysteriumnetwork/myst-launcher/controller/docker"
	"github.com/mysteriumnetwork/myst-launcher/controller/native"
	"github.com/mysteriumnetwork/myst-launcher/model"
)

func NewBackend(n string, m *model.UIModel, ui model.Gui_) model.RunnerController {
	switch n {
	case "native":
		return native.NewSvc(m, ui)
	case "docker":
		return docker.NewSvc(m, ui)
	default:
		return nil
	}
}
