package gui_win32

import (
	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
	. "github.com/mysteriumnetwork/myst-launcher/widget/declarative"
	"github.com/mysteriumnetwork/myst-launcher/widget/impl"

	"github.com/mysteriumnetwork/myst-launcher/model"
)

type InstallationFrame struct {
	*walk.Composite
	mdl *model.UIModel

	btnFinish            *walk.PushButton
	checkWindowsVersion  *impl.StatusViewImpl
	checkVirt            *impl.StatusViewImpl
	installExecutable    *impl.StatusViewImpl
	rebootAfterWSLEnable *impl.StatusViewImpl
	downloadFiles        *impl.StatusViewImpl
	installWSLUpdate     *impl.StatusViewImpl
	installDocker        *impl.StatusViewImpl
	checkGroupMembership *impl.StatusViewImpl
	lbInstallationStatus *walk.TextEdit
}

func NewInstallationFrame(parent walk.Container, mdl *model.UIModel) *InstallationFrame {
	f := new(InstallationFrame)
	f.mdl = mdl

	c := Composite{
		AssignTo: &f.Composite,

		Layout: VBox{
			MarginsZero: true,
		},

		Children: []Widget{
			GroupBox{
				Title:  "Installation process",
				Layout: Grid{Columns: 2},
				Children: []Widget{
					VSpacer{ColumnSpan: 2},
					VSeparator{ColumnSpan: 2}, // workaround

					Composite{
						Layout: HBox{
							MarginsZero: true,
						},
						Children: []Widget{
							StatusView{
								AssignTo: &f.checkWindowsVersion,
								MaxSize:  Size{Height: 20, Width: 20},
							},
							Label{
								Text: "Check Windows version",
							},
							HSpacer{},
						},
						MaxSize:    Size{Height: 20},
						ColumnSpan: 2,
					},
					Composite{
						Layout: HBox{
							MarginsZero: true,
						},
						Children: []Widget{
							StatusView{
								AssignTo: &f.installExecutable,
								MaxSize:  Size{Height: 20, Width: 20},
							},
							Label{
								Text: "Install executable",
							},
							HSpacer{},
						},
						MaxSize:    Size{Height: 20},
						ColumnSpan: 2,
					},

					Composite{
						Layout: HBox{
							MarginsZero: true,
						},
						Children: []Widget{
							StatusView{
								AssignTo: &f.checkVirt,
								MaxSize:  Size{Height: 20, Width: 20},
							},
							Label{
								Text: "Check Virtualization",
							},
							HSpacer{},
						},
						MaxSize:    Size{Height: 20},
						ColumnSpan: 2,
					},

					Composite{
						Layout: HBox{
							MarginsZero: true,
						},
						Children: []Widget{
							StatusView{
								AssignTo: &f.rebootAfterWSLEnable,
								MaxSize:  Size{Height: 20, Width: 20},
							},
							Label{
								Text: "Reboot after WSL enable",
							},
							HSpacer{},
						},
						MaxSize:    Size{Height: 20},
						ColumnSpan: 2,
					},

					Composite{
						Layout: HBox{
							MarginsZero: true,
						},
						Children: []Widget{
							StatusView{
								AssignTo: &f.downloadFiles,
								MaxSize:  Size{Height: 20, Width: 20},
							},
							Label{
								Text: "Download files",
							},
							HSpacer{},
						},
						MaxSize:    Size{Height: 20},
						ColumnSpan: 2,
					},

					Composite{
						Layout: HBox{
							MarginsZero: true,
						},
						Children: []Widget{
							StatusView{
								AssignTo: &f.installWSLUpdate,
								MaxSize:  Size{Height: 20, Width: 20},
							},
							Label{
								Text: "Install WSL update",
							},
							HSpacer{},
						},
						MaxSize:    Size{Height: 20},
						ColumnSpan: 2,
					},

					Composite{
						Layout: HBox{
							MarginsZero: true,
						},
						Children: []Widget{
							StatusView{
								AssignTo: &f.installDocker,
								MaxSize:  Size{Height: 20, Width: 20},
							},
							Label{
								Text: "Install Docker",
							},
							HSpacer{},
						},
						MaxSize:    Size{Height: 20},
						ColumnSpan: 2,
					},

					Composite{
						Layout: HBox{
							MarginsZero: true,
						},
						Children: []Widget{
							StatusView{
								AssignTo: &f.checkGroupMembership,
								MaxSize:  Size{Height: 20, Width: 20},
							},
							Label{
								Text: "Check group membership (docker-users)",
							},
							HSpacer{},
						},
						MaxSize:    Size{Height: 20},
						ColumnSpan: 2,
					},

					VSpacer{
						ColumnSpan: 2,
						MinSize: Size{
							Height: 24,
						},
					},
					Label{
						Text:       "Installation status:",
						ColumnSpan: 2,
					},

					TextEdit{
						ColumnSpan: 2,
						RowSpan:    1,
						AssignTo:   &f.lbInstallationStatus,
						ReadOnly:   true,
						MaxSize: Size{
							Height: 120,
						},
						VScroll: true,
					},

					VSpacer{ColumnSpan: 2, Row: 1},
					PushButton{
						ColumnSpan: 2,
						AssignTo:   &f.btnFinish,
						Text:       "Finish",
						OnClicked: func() {
							mdl.UIBus.Publish("click-finish")
						},
					},
				},
			},
		},
	}
	c.Create(NewBuilder(parent))

	f.mdl.UIBus.Subscribe("model-change", f.handlerState)
	mdl.UIBus.Subscribe("want-exit", f.handler)
	mdl.UIBus.Subscribe("log", f.handlerLog)

	return f
}

func (f *InstallationFrame) handler() {
	f.Synchronize(func() {
		f.btnFinish.SetEnabled(true)
	})
}

func (f *InstallationFrame) handlerState() {
	f.Synchronize(func() {
		switch f.mdl.State {
		case model.UIStateInitial:
		case model.UIStateInstallInProgress:
			f.btnFinish.SetEnabled(false)

		case model.UIStateInstallFinished:
			f.btnFinish.SetEnabled(true)
			f.btnFinish.SetText("Finish")

		case model.UIStateInstallError:
			f.btnFinish.SetEnabled(true)
			f.btnFinish.SetText("Exit installer")
		}

		switch f.mdl.State {
		case model.UIStateInstallInProgress, model.UIStateInstallFinished, model.UIStateInstallError:
			setState(f.checkWindowsVersion, f.mdl.CheckWindowsVersion)
			setState(f.checkVirt, f.mdl.CheckVirt)
			setState(f.installExecutable, f.mdl.InstallExecutable)
			setState(f.rebootAfterWSLEnable, f.mdl.RebootAfterWSLEnable)
			setState(f.downloadFiles, f.mdl.DownloadFiles)
			setState(f.installWSLUpdate, f.mdl.InstallWSLUpdate)
			setState(f.installDocker, f.mdl.InstallDocker)
			setState(f.checkGroupMembership, f.mdl.CheckGroupMembership)
		}
	})
}

func (f *InstallationFrame) handlerLog(p []byte) {
	f.Synchronize(func() {
		switch f.mdl.State {
		case model.UIStateInstallInProgress, model.UIStateInstallError, model.UIStateInstallFinished:
			f.Synchronize(func() {
				f.lbInstallationStatus.AppendText(string(p) + "\r\n")
			})
		}
	})
}

func (f *InstallationFrame) Close() {
	f.mdl.UIBus.Unsubscribe("model-change", f.handlerState)
	f.mdl.UIBus.Unsubscribe("want-exit", f.handler)
	f.mdl.UIBus.Unsubscribe("log", f.handlerLog)

	f.Dispose()
}
