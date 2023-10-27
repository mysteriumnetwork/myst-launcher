//go:build linux
// +build linux

package platform

type Manager struct{}

func NewManager() (*Manager, error) {
	m := &Manager{}
	return m, nil
}

func (m *Manager) Features() (bool, error) {
	return false, nil
}

func (m *Manager) SystemUnderVm() (bool, error) {
	return false, nil
}
