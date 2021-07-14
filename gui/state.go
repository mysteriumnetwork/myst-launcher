package gui

type modalState int

const (
	// state
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
