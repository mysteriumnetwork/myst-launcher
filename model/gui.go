package model

type Gui_ interface {
	CloseUI()

	DialogueComplete()
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
