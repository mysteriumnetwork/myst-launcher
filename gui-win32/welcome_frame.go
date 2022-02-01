package gui_win32

import (
	"github.com/gonutz/w32"
	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"

	"github.com/mysteriumnetwork/myst-launcher/model"
)

type Frame interface {
	Close()
}

type InstallationWelcomeFrame struct {
	*walk.Composite
	btnBegin *walk.PushButton
	mdl      *model.UIModel
}

func NewInstallationWelcomeFrame(parent walk.Container, mdl *model.UIModel) *InstallationWelcomeFrame {
	f := new(InstallationWelcomeFrame)
	f.mdl = mdl

	c := Composite{
		AssignTo: &f.Composite,

		Layout: VBox{
			MarginsZero: true,
		},

		Children: []Widget{
			GroupBox{
				Title:  "Installation",
				Layout: VBox{},
				Children: []Widget{
					HSpacer{ColumnSpan: 1},
					VSpacer{RowSpan: 1},
					Label{
						Text: "Installation status:",
					},
					TextEdit{
						Text: "This wizard will help with installation of missing components to run Mysterium Node.\r\n\r\n" +
							"Please press Install button to proceed with installation.",
						ReadOnly: true,
						MaxSize: Size{
							Height: 120,
						},
					},
					VSpacer{Row: 1},
					PushButton{
						AssignTo: &f.btnBegin,
						Text:     "Install",
						OnClicked: func() {
							//setState(1)
							mdl.UIBus.Publish("click-finish")
						},
					},
				},
			},
		},
	}
	c.Create(NewBuilder(parent))

	if !w32.SHIsUserAnAdmin() {
		f.btnBegin.SetImage(walk.IconShield())
	}
	mdl.UIBus.Subscribe("state-change", f.handlerState)

	return f
}

func (f *InstallationWelcomeFrame) handlerState() {
	f.Synchronize(func() {
		if f.mdl.State == model.UIStateInstallNeeded {
			f.btnBegin.SetEnabled(true)
		}
	})
}

func (f *InstallationWelcomeFrame) Close() {
	f.mdl.UIBus.Unsubscribe("state-change", f.handlerState)
	f.Dispose()
}
