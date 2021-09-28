package model

type UIState int

const (
	// model
	UIStateInitial           UIState = 0
	UIStateInstallNeeded     UIState = -1
	UIStateInstallInProgress UIState = -2
	UIStateInstallFinished   UIState = -3
	UIStateInstallError      UIState = -4
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
