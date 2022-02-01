/**
 * Copyright (c) 2021 BlockDev AG
 *
 * This source code is licensed under the MIT license found in the
 * LICENSE file in the root directory of this source tree.
 */

package gui_win32

import (
	"github.com/mysteriumnetwork/myst-launcher/utils"
	"log"

	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"

	"github.com/mysteriumnetwork/myst-launcher/model"
	. "github.com/mysteriumnetwork/myst-launcher/widget/declarative"
	"github.com/mysteriumnetwork/myst-launcher/widget/impl"
)

type StateFrame struct {
	*walk.Composite
	mdl *model.UIModel

	stDocker         *impl.StatusViewImpl
	stContainer      *impl.StatusViewImpl
	lbDocker         *walk.Label
	lbContainer      *walk.Label
	lbVersionCurrent *walk.Label
	lbVersionLatest  *walk.Label
	lbImageName      *walk.Label

	autoUpgrade   *walk.CheckBox
	lbNetworkMode *walk.Label
	btnOpenNodeUI *walk.PushButton

	lbNetwork  *walk.Label
	btnMainNet *walk.PushButton

	cmp              *walk.Composite
	headerContainer  *walk.Composite
	lbNodeUI         *walk.LinkLabel
	lbMMN            *walk.LinkLabel
	lbUpdateLauncher *walk.LinkLabel
	img              *walk.ImageView
}

func NewStateFrame(parent walk.Container, mdl *model.UIModel) *StateFrame {

	f := new(StateFrame)
	f.mdl = mdl

	c := Composite{
		AssignTo: &f.Composite,

		Layout: VBox{
			MarginsZero: true,
		},

		Children: []Widget{
			Composite{
				AssignTo: &f.cmp,
				MinSize:  Size{Height: 120},
				MaxSize:  Size{Height: 120},

				Children: []Widget{
					ImageView{
						AssignTo:  &f.img,
						Alignment: AlignHNearVFar,
					},

					Composite{
						AssignTo: &f.headerContainer,
						Layout: HBox{
							MarginsZero: true,
						},
						Children: []Widget{
							StatusView{
								AssignTo: &f.stContainer,
								MaxSize:  Size{Height: 20, Width: 20},
							},
							Label{
								Text:     "-",
								AssignTo: &f.lbContainer,
								Font: Font{
									PointSize: 8,
									Bold:      true,
								},
							},
							LinkLabel{
								Text:     "<a>Node UI</a>",
								AssignTo: &f.lbNodeUI,
								OnLinkActivated: func(link *walk.LinkLabelLink) {
									OpenNodeUI()
								},
								Alignment: AlignHNearVNear,
							},
							LinkLabel{
								Text:     "<a>Mystnodes.com</a>",
								AssignTo: &f.lbMMN,
								OnLinkActivated: func(link *walk.LinkLabelLink) {
									OpenMMN()
								},
								Alignment: AlignHNearVNear,
							},
							LinkLabel{
								Text:     "<a>Launcher update available</a>",
								AssignTo: &f.lbUpdateLauncher,
								OnLinkActivated: func(link *walk.LinkLabelLink) {
									mdl.UIBus.Publish("launcher-update")
								},
								Alignment: AlignHNearVNear,
								TextColor: walk.RGB(255, 55, 95),
							},
						},
						MaxSize: Size{Height: 20},
					},
				},
			},

			Composite{
				Layout: Grid{Columns: 2},
				Children: []Widget{
					VSpacer{ColumnSpan: 2},
					VSpacer{ColumnSpan: 2},

					Label{
						Text:     "Node info",
						AssignTo: &f.lbDocker,
						Font: Font{
							PointSize: 8,
							Bold:      true,
						},
					},
					Label{
						Text: "",
					},

					Label{
						Text: "Network",
					},
					Composite{
						Alignment: AlignHNearVCenter,
						Layout: HBox{
							MarginsZero: true,
							Spacing:     0,
						},

						Children: []Widget{
							Label{
								Text:      "-",
								AssignTo:  &f.lbNetwork,
								Alignment: AlignHNearVCenter,
							},
							PushButton{
								Alignment: AlignHNearVCenter,
								Enabled:   true,
								AssignTo:  &f.btnMainNet,
								Text:      "Update to MainNet..",

								OnSizeChanged: func() {
									f.btnMainNet.SetWidthPixels(115)
								},
								OnClicked: func() {
									mdl.UIBus.Publish("btn-upgrade-network")
								},
							},
							//HSpacer{StretchFactor: 0},
						},
					},
					Label{
						Text: "",
					},
					LinkLabel{
						Text: "<a>Information about MainNet</a>",
						OnLinkActivated: func(link *walk.LinkLabelLink) {
							utils.OpenUrlInBrowser("https://mysterium.network/")
						},
						Alignment: AlignHNearVNear,
					},
					VSpacer{ColumnSpan: 2, Size: 10},

					Label{
						Text: "Docker Hub image name",
					},
					Label{
						AssignTo: &f.lbImageName,
					},

					Label{
						Text: "Node version installed",
					},
					Label{
						Text:     "-",
						AssignTo: &f.lbVersionCurrent,
					},
					Label{
						Text: "Latest version",
					},
					Label{
						Text:     "-",
						AssignTo: &f.lbVersionLatest,
					},

					Label{
						Text: "Upgrade automatically",
					},
					CheckBox{
						Text:     " ",
						AssignTo: &f.autoUpgrade,
						OnCheckedChanged: func() {
							mdl.GetConfig().AutoUpgrade = f.autoUpgrade.Checked()
							mdl.GetConfig().Save()
						},
						MaxSize: Size{Height: 15},
					},

					VSpacer{
						ColumnSpan: 2,
						Size:       20,
					},

					Label{
						Text:     "Advanced settings",
						AssignTo: &f.lbDocker,
						Font: Font{
							PointSize: 8,
							Bold:      true,
						},
					},
					Label{
						Text: "",
					},
					Label{
						Text: "Networking mode",
					},
					Label{
						Text:     "-",
						AssignTo: &f.lbNetworkMode,
					},
					Label{
						Text: "",
					},
					PushButton{
						Enabled:  true,
						AssignTo: &f.btnOpenNodeUI,
						Text:     "Config..",
						OnSizeChanged: func() {
							f.btnOpenNodeUI.SetWidthPixels(75)
						},
						OnClicked: func() {
							mdl.UIBus.Publish("btn-config-click")
						},
					},
					VSpacer{
						ColumnSpan: 2,
						Size:       20,
					},

					Label{
						Text: "Docker Desktop",
						Font: Font{
							PointSize: 8,
							Bold:      true,
						},
					},
					Label{
						Text: "",
					},
					Label{
						Text: "Status",
					},

					Composite{
						Layout: HBox{
							MarginsZero: true,
							//Margins:   Margins{},
							Alignment: AlignHNearVNear,
							Spacing:   0,
						},
						Children: []Widget{
							StatusView{
								AssignTo: &f.stDocker,
								MaxSize:  Size{Height: 20, Width: 20},
							},
							Label{
								Accessibility: Accessibility{
									Role: AccRoleStatictext,
								},
								Text:     "-",
								AssignTo: &f.lbDocker,
								Font: Font{
									PointSize: 8,
									Bold:      true,
								},
							},
						},
						MaxSize: Size{Height: 20},
					},

					VSpacer{
						ColumnSpan: 2,
						Size:       20,
					},

					VSpacer{ColumnSpan: 2},
				},
			},
		},
	}

	c.Create(NewBuilder(parent))
	f.btnOpenNodeUI.SetFocus()

	logo, err := walk.NewIconFromResourceWithSize("APPICON", walk.Size{64, 64})
	if err != nil {
		log.Fatal(err)
	}
	img, err := walk.ImageFrom(logo)
	if err != nil {
		log.Fatal(err)
	}
	f.img.SetImage(img)

	f.mdl.UIBus.Subscribe("model-change", f.handlerState)

	f.cmp.SetBounds(walk.Rectangle{0, 0, 400, 100})
	f.img.SetBounds(walk.Rectangle{5, 5, 70, 70})
	f.headerContainer.SetBounds(walk.Rectangle{80, 5, 270, 70})
	f.stContainer.SetBounds(walk.Rectangle{0, 0, 20, 20})
	f.lbContainer.SetBounds(walk.Rectangle{25, 0, 70, 20})
	f.lbNodeUI.SetBounds(walk.Rectangle{120, 4, 150, 20})
	f.lbMMN.SetBounds(walk.Rectangle{120, 24, 150, 20})
	f.lbUpdateLauncher.SetBounds(walk.Rectangle{120, 44, 150, 20})

	return f
}

func (f *StateFrame) handlerState() {
	//fmt.Println("handlerState >>>>>>>>>>>>>>>>>>>>>>>>", f.mdl.State, f.mdl.StateContainer)

	f.Synchronize(func() {

		switch f.mdl.State {
		case model.UIStateInitial:
			f.autoUpgrade.SetChecked(f.mdl.GetConfig().AutoUpgrade)
			if !f.mdl.GetConfig().EnablePortForwarding {
				f.lbNetworkMode.SetText(`Port restricted cone NAT`)
			} else {
				f.lbNetworkMode.SetText(`Manual port forwarding`)
			}

			f.lbDocker.SetText(f.mdl.StateDocker.String())
			f.lbContainer.SetText(f.mdl.StateContainer.String())
			setState2(f.stDocker, f.mdl.StateDocker)
			setState2(f.stContainer, f.mdl.StateContainer)

			f.lbContainer.SetText(f.mdl.StateContainer.String())
			if !f.mdl.GetConfig().Enabled {
				f.lbContainer.SetText("Disabled")
			}
			f.btnOpenNodeUI.SetEnabled(f.mdl.IsRunning())

			f.lbVersionCurrent.SetText(f.mdl.ImageInfo.VersionCurrent)
			f.lbVersionLatest.SetText(f.mdl.ImageInfo.VersionLatest)

			f.lbImageName.SetText(f.mdl.Config.GetFullImageName())
			f.btnOpenNodeUI.SetFocus()

			f.lbNetwork.SetText(f.mdl.Config.GetNetworkCaption())
			f.btnMainNet.SetVisible(!f.mdl.CurrentNetIsMainNet())
			f.lbUpdateLauncher.SetVisible(f.mdl.LauncherHasUpdate)
		}
	})
}

func (f *StateFrame) Close() {
	f.mdl.UIBus.Unsubscribe("model-change", f.handlerState)
	f.Dispose()
}

func OpenNodeUI() {
	utils.OpenUrlInBrowser("http://localhost:4449/")
}

func OpenMMN() {
	utils.OpenUrlInBrowser("https://mystnodes.com/")
}

func setState(b *impl.StatusViewImpl, st model.InstallStep) {
	switch st {
	case model.StepInProgress:
		b.SetState(3)
	case model.StepFinished:
		b.SetState(2)
	case model.StepFailed:
		b.SetState(1)
	case model.StepNone:
		b.SetState(0)
	default:
		b.SetState(0)
	}
}
func setState2(b *impl.StatusViewImpl, st model.RunnableState) {
	switch st {
	case model.RunnableStateRunning:
		b.SetState(2)
	case model.RunnableStateInstalling:
		b.SetState(3)
	case model.RunnableStateStarting:
		b.SetState(3)
	case model.RunnableStateUnknown:
		b.SetState(1)
	}
}
