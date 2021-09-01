package gui

type Gui_ interface {
	//OpenUpgradeDlg()
	//CreateNotifyIcon(ui *UIModel)
	//NetworkingDlg()
	//CreateMainWindow()
	//Run()

	ShowMain()
	ShowNotificationInstalled()
	ShowNotificationUpgrade()
	ConfirmModal(title, message string) int
	YesNoModal(title, message string) int
	ErrorModal(title, message string) int
}
