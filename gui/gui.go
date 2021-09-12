package gui

type Gui_ interface {
	//OpenUpgradeDlg()
	//CreateNotifyIcon(ui *UIModel)
	//NetworkingDlg()
	//CreateMainWindow()
	//Run()

	NotifyUIExitApp()
	WaitDialogueComplete() bool
	TerminateWaitDialogueComplete()

	ShowMain()
	ShowNotificationInstalled()
	ShowNotificationUpgrade()

	ConfirmModal(title, message string) int
	YesNoModal(title, message string) int
	ErrorModal(title, message string) int
	SetModalReturnCode(rc int)
}
