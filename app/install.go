package app

import (
	"github.com/mysteriumnetwork/myst-launcher/model"
)

type step struct {
	name   string
	action func() bool
}

type StepExec struct {
	s     *AppState
	steps []step
}

func (e *StepExec) AddStep(stepName string, f func() bool) {
	e.steps = append(e.steps, step{
		name:   stepName,
		action: f,
	})
}

func (e *StepExec) Run() bool {
	for _, step := range e.steps {
		e.s.model.UpdateProperties(model.UIProps{step.name: model.StepInProgress})
		if !step.action() {
			e.s.model.UpdateProperties(model.UIProps{step.name: model.StepFailed})
			e.s.model.SwitchState(model.UIStateInstallError)

			return false
		}
		e.s.model.UpdateProperties(model.UIProps{step.name: model.StepFinished})
	}
	return true
}
