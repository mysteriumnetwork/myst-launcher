package gui

import (
	. "github.com/lxn/walk/declarative"
)

func (mw *Gui) installationWelcome() Widget {
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
						AssignTo: &mw.btnBegin,
						Text:     "Install",
						OnClicked: func() {
							UI.BtnFinishOnClick()
						},
					},
				},
			},
		},
	}
}

func (mw *Gui) installationDlg() Widget {
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
					VSpacer{RowSpan: 1, ColumnSpan: 2},
					Label{
						Text: "Check Windows version",
					},
					CheckBox{
						Enabled:  false,
						AssignTo: &mw.checkWindowsVersion,
					},

					Label{
						Text: "Check VT-x",
					},
					CheckBox{
						Enabled:  false,
						AssignTo: &mw.checkVTx,
					},
					Label{
						Text: "Check WSL",
					},
					CheckBox{
						Enabled:  false,
						AssignTo: &mw.enableWSL,
					},
					Label{
						Text: "Check Hyper-V",
					},
					CheckBox{
						Enabled:  false,
						AssignTo: &mw.enableHyperV,
					},

					Label{
						Text: "Install executable",
					},
					CheckBox{
						Enabled:  false,
						AssignTo: &mw.installExecutable,
					},
					Label{
						Text: "Reboot after WSL enable",
					},
					CheckBox{
						Enabled:  false,
						AssignTo: &mw.rebootAfterWSLEnable,
					},
					Label{
						Text: "Download files",
					},
					CheckBox{
						Enabled:  false,
						AssignTo: &mw.downloadFiles,
					},
					Label{
						Text: "Install WSL update",
					},
					CheckBox{
						Enabled:  false,
						AssignTo: &mw.installWSLUpdate,
					},
					Label{
						Text: "Install Docker",
					},
					CheckBox{
						Enabled:  false,
						AssignTo: &mw.installDocker,
					},
					Label{
						Text: "Check group membership (docker-users)",
					},
					CheckBox{
						Enabled:  false,
						AssignTo: &mw.checkGroupMembership,
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
						AssignTo:   &mw.lbInstallationStatus,
						ReadOnly:   true,
						MaxSize: Size{
							Height: 120,
						},
						VScroll: true,
					},

					VSpacer{ColumnSpan: 2, Row: 1},
					PushButton{
						ColumnSpan: 2,
						AssignTo:   &mw.btnFinish,
						Text:       "Finish",
						OnClicked: func() {
							UI.BtnFinishOnClick()
						},
					},
				},
			},
		},
	}
}

func (mw *Gui) stateDlg() Widget {
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
						Text: "Current node version",
					},
					Label{
						Text:     "-",
						AssignTo: &mw.lbVersionCurrent,
					},
					Label{
						Text: "Latest node version",
					},
					Label{
						Text:     "-",
						AssignTo: &mw.lbVersionLatest,
					},
					Label{
						Text: "Docker Hub image name",
					},
					Label{
						Text: UI.app.GetImageName(),
					},
					Label{
						Text:       "-",
						ColumnSpan: 2,
					},

					Label{
						Text: "Docker",
					},
					Label{
						Text:     "-",
						AssignTo: &mw.lbDocker,
					},
					Label{
						Text: "Container",
					},
					Label{
						Text:     "-",
						AssignTo: &mw.lbContainer,
					},
					CheckBox{
						Text:     "Start automatically",
						AssignTo: &mw.autoStart,
						OnCheckedChanged: func() {
							UI.app.GetConfig().AutoStart = mw.autoStart.Checked()
							UI.app.SaveConfig()
						},
						ColumnSpan: 2,
					},
					PushButton{
						Enabled:  false,
						AssignTo: &mw.btnOpenNodeUI,
						Text:     "Open Node UI",
						OnClicked: func() {
							UI.OpenNodeUI()
						},
						ColumnSpan: 2,
					},
					VSpacer{ColumnSpan: 2},
				},
			},
		},
	}
}
