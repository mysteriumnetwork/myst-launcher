/**
 * Copyright (c) 2021 BlockDev AG
 *
 * This source code is licensed under the MIT license found in the
 * LICENSE file in the root directory of this source tree.
 */

package gui_win32

import (
	"log"

	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
	_const "github.com/mysteriumnetwork/myst-launcher/const"

	"github.com/mysteriumnetwork/myst-launcher/model"
	"github.com/mysteriumnetwork/myst-launcher/utils"
	. "github.com/mysteriumnetwork/myst-launcher/widget/declarative"
	"github.com/mysteriumnetwork/myst-launcher/widget/impl"
)

type StateFrame struct {
	*walk.Composite
	mdl *model.UIModel

	stDocker              *impl.StatusViewImpl
	stContainer           *impl.StatusViewImpl
	lbBackend             *walk.Label
	lbVendorID            *walk.Label
	lbDockerDesktopStatus *walk.Label
	lbDocker              *walk.Label
	lbContainer           *walk.Label
	lbVersionCurrent      *walk.Label
	lbVersionLatest       *walk.Label
	lbImageName           *walk.Label

	autoUpgrade       *walk.CheckBox
	lbNetworkMode     *walk.Label
	btnOpenNodeConfig *walk.PushButton

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
									mdl.UIBus.Publish("launcher-trigger-update")
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
						Text: "Node package source",
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
						AssignTo: &f.btnOpenNodeConfig,
						Text:     "Config..",
						OnSizeChanged: func() {
							f.btnOpenNodeConfig.SetWidthPixels(75)
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
						Text: "Backend",
						Font: Font{
							PointSize: 8,
							Bold:      true,
						},
					},
					Label{
						Text:     "",
						AssignTo: &f.lbBackend,
					},
					Label{
						Text: "Vendor Id",
						Font: Font{
							PointSize: 8,
							Bold:      true,
						},
					},
					Label{
						Text:     "",
						AssignTo: &f.lbVendorID,
					},

					Label{
						Text:     "Status",
						AssignTo: &f.lbDockerDesktopStatus,
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
	f.btnOpenNodeConfig.SetFocus()

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

			isNative := f.mdl.Config.Backend == "native"
			{
				f.lbBackend.SetText("native")
				if !isNative {
					f.lbBackend.SetText("docker")
				}

				f.lbDockerDesktopStatus.SetVisible(!isNative)
				f.stDocker.SetVisible(!isNative)
				f.lbDocker.SetVisible(!isNative)
			}

			f.lbVendorID.SetText(_const.VendorID)
			if f.lbVendorID.Text() == "" {
				f.lbVendorID.SetText("-")
			}

			f.lbDocker.SetText(f.mdl.StateDocker.String())
			f.lbContainer.SetText(f.mdl.StateContainer.String())
			setState2(f.stDocker, f.mdl.StateDocker)
			setState2(f.stContainer, f.mdl.StateContainer)

			f.lbContainer.SetText(f.mdl.StateContainer.String())
			f.btnOpenNodeConfig.SetEnabled(f.mdl.IsRunning() && f.mdl.Caps == 1)

			f.lbVersionCurrent.SetText(f.mdl.ImageInfo.VersionCurrent)
			f.lbVersionLatest.SetText(f.mdl.ImageInfo.VersionLatest)

			f.lbImageName.SetText(f.mdl.Config.GetFullImageName())

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
