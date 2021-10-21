package gui_win32

import (
	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
	. "github.com/mysteriumnetwork/myst-launcher/widget/declarative"
)

func (g *Gui) installationWelcome() Widget {
	return Composite{
		Visible: false,
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
						AssignTo: &g.btnBegin,
						Text:     "Install",
						OnClicked: func() {
							if g.model.WantExit {
								g.CloseUI()
							}
							g.DialogueComplete()
						},
					},
				},
			},
		},
	}
}

func (g *Gui) installationDlg() Widget {
	return Composite{
		Visible: false,
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
								AssignTo: &g.checkWindowsVersion,
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
								AssignTo: &g.installExecutable,
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
								AssignTo: &g.checkVirt,
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
								AssignTo: &g.rebootAfterWSLEnable,
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
								AssignTo: &g.downloadFiles,
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
								AssignTo: &g.installWSLUpdate,
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
								AssignTo: &g.installDocker,
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
								AssignTo: &g.checkGroupMembership,
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
						AssignTo:   &g.lbInstallationStatus,
						ReadOnly:   true,
						MaxSize: Size{
							Height: 120,
						},
						VScroll: true,
					},

					VSpacer{ColumnSpan: 2, Row: 1},
					PushButton{
						ColumnSpan: 2,
						AssignTo:   &g.btnFinish,
						Text:       "Finish",
						OnClicked: func() {
							if g.model.WantExit {
								g.CloseUI()
							}
							g.DialogueComplete()
						},
					},
				},
			},
		},
	}
}

func (g *Gui) stateDlg() Widget {
	return Composite{
		Name:    "HH",
		Visible: false,
		Layout: VBox{
			MarginsZero: true,
		},

		Children: []Widget{
			Composite{
				AssignTo: &g.cmp,
				MinSize:  Size{Height: 120},
				MaxSize:  Size{Height: 120},

				Children: []Widget{
					ImageView{
						AssignTo:  &g.img,
						Alignment: AlignHNearVFar,
					},

					Composite{
						AssignTo: &g.headerContainer,
						Layout: HBox{
							MarginsZero: true,
						},
						Children: []Widget{
							StatusView{
								AssignTo: &g.stContainer,
								MaxSize:  Size{Height: 20, Width: 20},
							},
							Label{
								Text:     "-",
								AssignTo: &g.lbContainer,
								Font: Font{
									PointSize: 8,
									Bold:      true,
								},
							},

							LinkLabel{
								Text:     "<a>Node UI</a>",
								AssignTo: &g.lbNodeUI,
								OnLinkActivated: func(link *walk.LinkLabelLink) {
									OpenNodeUI()
								},
								Alignment: AlignHNearVNear,
							},
							LinkLabel{
								Text:     "<a>My Mysterium Network</a>",
								AssignTo: &g.lbMMN,
								OnLinkActivated: func(link *walk.LinkLabelLink) {
									OpenMMN()
								},
								Alignment: AlignHNearVNear,
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
						AssignTo: &g.lbDocker,
						Font: Font{
							PointSize: 8,
							Bold:      true,
						},
					},
					Label{
						Text: "",
					},
					Label{
						Text: "Docker Hub image name",
					},
					Label{
						AssignTo: &g.lbImageName,
					},

					Label{
						Text: "Node version installed",
					},
					Label{
						Text:     "-",
						AssignTo: &g.lbVersionCurrent,
					},
					Label{
						Text: "Latest version",
					},
					Label{
						Text:     "-",
						AssignTo: &g.lbVersionLatest,
					},

					Label{
						Text: "Upgrade automatically",
					},
					CheckBox{
						Text:     " ",
						AssignTo: &g.autoUpgrade,
						OnCheckedChanged: func() {
							g.model.GetConfig().AutoUpgrade = g.autoUpgrade.Checked()
							g.model.GetConfig().Save()
						},
						MaxSize: Size{Height: 15},
					},

					VSpacer{
						ColumnSpan: 2,
						Size:       20,
					},

					Label{
						Text:     "Advanced settings",
						AssignTo: &g.lbDocker,
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
						AssignTo: &g.lbNetworkMode,
					},
					Label{
						Text: "",
					},
					PushButton{
						Enabled:  true,
						AssignTo: &g.btnOpenNodeUI,
						Text:     "Config..",

						OnSizeChanged: func() {
							g.btnOpenNodeUI.SetWidthPixels(75)
						},						
						OnClicked: func() {
							g.NetworkingDlg()
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
						},
						Children: []Widget{
							StatusView{
								AssignTo:        &g.stDocker,
								MaxSize:         Size{Height: 20, Width: 20},
								OnBoundsChanged: func() {},
							},
							Label{
								Text:     "-",
								AssignTo: &g.lbDocker,
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

}
