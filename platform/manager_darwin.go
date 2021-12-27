package platform

import (
	"bytes"
	"log"
	"strings"

	"github.com/mysteriumnetwork/myst-launcher/utils"
)

const (
	FeatureHyperisorFramework = "HyperisorFramework"
)

var features = []string{
	FeatureHyperisorFramework,
}

type Manager struct{}

func NewManager() (*Manager, error) {
	m := &Manager{}
	return m, nil
}

func (m *Manager) Features() (bool, error) {

	hasHV := func() bool {
		var out bytes.Buffer
		_, err := utils.CmdRun(&out, "sysctl", "machdep.cpu.features")
		if err != nil {
			log.Println("QueryFeatures >", err)
			return false
		}
		if !strings.Contains(out.String(), "VMX") {
			return false
		}
		out.Reset()

		_, err = utils.CmdRun(&out, "sysctl", "kern.hv_support")
		if err != nil {
			log.Println("QueryFeatures >", err)
			return false
		}
		if !strings.HasPrefix(out.String(), "kern.hv_support: 1") {
			return false
		}
		return true
	}()
	if !hasHV {
		return false, nil
	}
	return true, nil

}

func (m *Manager) SystemUnderVm() (bool, error) {
	log.Println("SystemUnderVm: not implemented")
	return false, nil
}
