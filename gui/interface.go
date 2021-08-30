package gui

// Dialog box command ids
const (
	IDOK       = 1
	IDCANCEL   = 2
	IDABORT    = 3
	IDRETRY    = 4
	IDIGNORE   = 5
	IDYES      = 6
	IDNO       = 7
	IDCLOSE    = 8
	IDHELP     = 9
	IDTRYAGAIN = 10
	IDCONTINUE = 11
	IDTIMEOUT  = 32000
)

type UIProps map[string]interface{}

type UIModelInterface interface {
	WaitDialogueComplete() bool
	SetStateDocker(RunnableState)
	SetStateContainer(RunnableState)
	SwitchState(s ModalState)

	//ErrorModal(string, string) int
	//YesNoModal(string, string) int
	//ConfirmModal(string, string) int

	UpdateProperties(m UIProps)
	Update()
	ExitApp()
	SetWantExit()

	//ShowNotificationInstalled()

	Publish(topic string, args ...interface{})
}
