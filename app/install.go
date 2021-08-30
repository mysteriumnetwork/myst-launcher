// +build !windows

package app

// returns exit model: true means exit
func (s *AppState) tryInstall() bool {
	return false
}
