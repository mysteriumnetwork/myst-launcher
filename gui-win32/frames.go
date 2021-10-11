package gui_win32

import (
	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
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
					
					Label{
						Text: "Check Windows version",
					},
					CheckBox{
						Enabled:  false,
						AssignTo: &g.checkWindowsVersion,
					},

					Label{
						Text: "Install executable",
					},
					CheckBox{
						Enabled:  false,
						AssignTo: &g.installExecutable,
					},

					Label{
						Text: "Check Virtualization",
					},
					CheckBox{
						Enabled:  false,
						AssignTo: &g.checkVirt,
					},

					Label{
						Text: "Reboot after WSL enable",
					},
					CheckBox{
						Enabled:  false,
						AssignTo: &g.rebootAfterWSLEnable,
					},
					Label{
						Text: "Download files",
					},
					CheckBox{
						Enabled:  false,
						AssignTo: &g.downloadFiles,
					},
					Label{
						Text: "Install WSL update",
					},
					CheckBox{
						Enabled:  false,
						AssignTo: &g.installWSLUpdate,
					},
					Label{
						Text: "Install Docker",
					},
					CheckBox{
						Enabled:  false,
						AssignTo: &g.installDocker,
					},
					Label{
						Text: "Check group membership (docker-users)",
					},
					CheckBox{
						Enabled:  false,
						AssignTo: &g.checkGroupMembership,
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
		Visible: false,
		Layout: VBox{
			MarginsZero: true,
		},

		Children: []Widget{
			GroupBox{
				Title:  "Status",
				Layout: Grid{Columns: 2},
				Children: []Widget{
					VSpacer{ColumnSpan: 2},
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
						Text: "Upgrade available",
					},
					LinkLabel{
						AssignTo: &g.lbVersionUpdatesAvail,
						Text:     `-`,
						OnLinkActivated: func(link *walk.LinkLabelLink) {
							g.OpenUpgradeDlg()
						},
					},

					CheckBox{
						Text:           "Upgrade automatically",
						TextOnLeftSide: true,
						AssignTo:       &g.autoUpgrade,
						OnCheckedChanged: func() {
							g.model.GetConfig().AutoUpgrade = g.autoUpgrade.Checked()
							g.model.GetConfig().Save()
						},
						//ColumnSpan: 2,
						MaxSize: Size{Height: 15},
					},
					HSpacer{Size: 1},

					VSpacer{
						ColumnSpan: 2,
						Size:       20,
					},

					Label{
						Text: "Docker",
					},
					Label{
						Text:     "-",
						AssignTo: &g.lbDocker,
					},
					Label{
						Text: "Node",
					},
					Label{
						Text:     "-",
						AssignTo: &g.lbContainer,
					},
					VSpacer{
						ColumnSpan: 2,
						Size:       20,
					},

					Label{
						Text: "Networking mode",
					},
					LinkLabel{
						Text:     "-",
						AssignTo: &g.lbNetworkMode,
						OnLinkActivated: func(link *walk.LinkLabelLink) {
							g.NetworkingDlg()
						},
					},
					VSpacer{
						ColumnSpan: 2,
						Size:       20,
					},

					PushButton{
						Enabled:  false,
						AssignTo: &g.btnOpenNodeUI,
						Text:     "Open Node UI",
						OnClicked: func() {
							OpenNodeUI()
						},
						ColumnSpan: 2,
					},

					VSpacer{ColumnSpan: 2},
					HSpacer{},
				},
			},
		},
	}
}
