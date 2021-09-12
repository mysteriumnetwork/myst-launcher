package gui

import "github.com/mysteriumnetwork/myst-launcher/model"

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

type Model interface {
	WaitDialogueComplete() bool
	ExitApp()
	SetWantExit()


	SetStateDocker(RunnableState)
	SetStateContainer(RunnableState)
	OnConfigRead()

	SwitchState(s ModalState)

	UpdateProperties(m UIProps)
	Update()
	
	Publish(topic string, args ...interface{})
	GetConfig() *model.Config
}
