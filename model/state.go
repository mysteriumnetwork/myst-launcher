package model

type modalState int

const (
	// model
	ModalStateInitial           modalState = 0
	ModalStateInstallNeeded     modalState = -1
	ModalStateInstallInProgress modalState = -2
	ModalStateInstallFinished   modalState = -3
	ModalStateInstallError      modalState = -4
)

type runnableState int

const (
	RunnableStateUnknown    runnableState = 0
	RunnableStateStarting   runnableState = 1
	RunnableStateRunning    runnableState = 2
	RunnableStateInstalling runnableState = 3
)

func (r runnableState) String() string {
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
