package gui

type modalState int

const (
	// model
	ModalStateInitial           modalState = 0
	ModalStateInstallNeeded     modalState = -1
	ModalStateInstallInProgress modalState = -2
	ModalStateInstallFinished   modalState = -3
	ModalStateInstallError      modalState = -4
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
