package docker

import (
	"github.com/mysteriumnetwork/myst-launcher/model"
)

type step struct {
	name   string
	action func() bool
}

type StepExec struct {
	model *model.UIModel
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
		e.model.UpdateProperties(model.UIProps{step.name: model.StepInProgress})
		if !step.action() {
			e.model.UpdateProperties(model.UIProps{step.name: model.StepFailed})
			e.model.SwitchState(model.UIStateInstallError)

			return false
		}
		e.model.UpdateProperties(model.UIProps{step.name: model.StepFinished})
	}
	return true
}
