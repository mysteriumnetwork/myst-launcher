package gui

type ModalState int

const (
	// model
	ModalStateInitial           ModalState = 0
	ModalStateInstallNeeded     ModalState = -1
	ModalStateInstallInProgress ModalState = -2
	ModalStateInstallFinished   ModalState = -3
	ModalStateInstallError      ModalState = -4
)

type RunnableState int

const (
	RunnableStateUnknown    RunnableState = 0
	RunnableStateStarting   RunnableState = 1
	RunnableStateRunning    RunnableState = 2
	RunnableStateInstalling RunnableState = 3
)

func (r RunnableState) String() string {
	switch r {
	case RunnableStateRunning:
		return "Running [OK]"
	case RunnableStateInstalling:
		return "Installing.."
	case RunnableStateStarting:
		return "Starting.."
	case RunnableStateUnknown:
		return "-"
	default:
		return "?"
	}
}
