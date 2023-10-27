//go:build linux
// +build linux

package controller

import (
	"github.com/mysteriumnetwork/myst-launcher/controller/native"
	"github.com/mysteriumnetwork/myst-launcher/model"
)

func NewController(n string) model.Controller {
	switch n {
	case "native":
		return native.NewController()
	default:
		return nil
	}
}
