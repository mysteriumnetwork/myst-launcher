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
		return "ONLINE"
	case RunnableStateInstalling:
		return "INSTALLING.."
	case RunnableStateStarting:
		return "STARTING.."
	case RunnableStateUnknown:
		return "OFFLINE"
	default:
		return "???"
	}
}

type InstallStep int

const (
	StepNone       = InstallStep(0)
	StepInProgress = InstallStep(1)
	StepFinished   = InstallStep(2)
	StepFailed     = InstallStep(3)
)
