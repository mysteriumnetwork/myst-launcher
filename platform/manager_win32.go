//go:build windows
// +build windows

package platform

import (
	"fmt"
	"log"
	"strings"

	"github.com/gabriel-samfira/go-wmi/wmi"
	"github.com/google/glazier/go/dism"
	"github.com/pkg/errors"
	"github.com/scjalliance/comshim"
	"golang.org/x/sys/windows"
)

const (
	FeatureWSL        = "Microsoft-Windows-Subsystem-Linux"
	FeatureHyperV     = "Microsoft-Hyper-V"
	FeatureVMPlatform = "VirtualMachinePlatform"
)

var features = []string{
	FeatureWSL,
	FeatureHyperV,
	FeatureVMPlatform,
}

type Manager struct {
	hasDism bool
	ses     dism.Session
}

func NewManager() (*Manager, error) {
	m := &Manager{}
	return m, nil
}

func wmiAPIInit() (*wmi.WMI, error) {
	comshim.Add(1)
	return wmi.NewConnection(".", `root\cimv2`)
}

func wmiAPIRelease(w *wmi.WMI) {
	w.Close()
	comshim.Done()
}

func (m *Manager) queryOptionalFeature(con *wmi.WMI, feature string) (error, bool) {
	log.Println("Query optional feature:", feature)
	res, err := con.CallMethod("ExecQuery", fmt.Sprintf("SELECT * FROM Win32_OptionalFeature Where Name='%s'", feature))
	if err != nil {
		return errors.Wrap(err, "ExecQuery"), false
	}

	el, err := res.ItemAtIndex(0)
	// err -> empty set
	if err != nil {
		return nil, false
	}

	state_, err := el.GetProperty("InstallState")
	if err != nil {
		return errors.Wrap(err, "GetProperty"), false
	}
	state := state_.Value().(int32)
	return nil, state == 1
}

func (m *Manager) Features() (bool, error) {
	w, _ := wmiAPIInit()
	defer wmiAPIRelease(w)

	for _, f := range features {
		err, installed := m.queryOptionalFeature(w, f)
		if err != nil {
			log.Println("Features >", err)
			return false, err
		}
		if !installed {
			log.Println("Feature not installed:", f)
			return false, nil
		}
	}
	return true, nil
}

func (m *Manager) SystemUnderVm() (bool, error) {
	con, err := wmiAPIInit()
	defer wmiAPIRelease(con)

	res, err := con.CallMethod("ExecQuery", "SELECT * FROM Win32_ComputerSystem")
	if err != nil {
		return false, errors.Wrap(err, "ExecQuery")
	}
	item, err := res.ItemAtIndex(0)
	if err != nil {
		return false, errors.Wrap(err, "ItemAtIndex")
	}
	prop_, err := item.GetProperty("Model")
	if err != nil {
		return false, errors.Wrap(err, "GetProperty")
	}
	model := prop_.Raw().ToString()

	vmTest := []string{"virtual", "vmware", "kvm", "xen"}
	isVM := false
	for _, v := range vmTest {
		if strings.Contains(strings.ToLower(model), v) {
			isVM = true
			break
		}
	}
	return isVM, nil
}

/*
 *  Win32-specific methods
 */

func (m *Manager) IsVMcomputeRunning() (bool, error) {
	con, _ := wmiAPIInit()
	defer wmiAPIRelease(con)

	res, err := con.CallMethod("ExecQuery", "SELECT * FROM Win32_Service Where Name='vmcompute'")
	if err != nil {
		return false, errors.Wrap(err, "ExecQuery")
	}
	item, err := res.ItemAtIndex(0)
	if err != nil {
		return false, errors.Wrap(err, "ItemAtIndex")
	}
	prop_, err := item.GetProperty("State")
	if err != nil {
		return false, errors.Wrap(err, "GetProperty")
	}
	state := prop_.Raw().ToString()
	return state == "Running", nil
}

func (m *Manager) StartVmcomputeIfNotRunning() (bool, error) {
	con, _ := wmiAPIInit()
	defer wmiAPIRelease(con)

	res, err := con.CallMethod("ExecQuery", "SELECT * FROM Win32_Service Where Name='vmcompute'")
	if err != nil {
		return false, errors.Wrap(err, "ExecQuery")
	}
	item, err := res.ItemAtIndex(0)
	if err != nil {
		return false, errors.Wrap(err, "ItemAtIndex")
	}
	prop_, err := item.GetProperty("State")
	if err != nil {
		return false, errors.Wrap(err, "GetProperty")
	}
	state := prop_.Raw().ToString()
	if state == "Running" {
		return true, nil
	}

	res, err = item.Get("StartService")
	if err != nil {
		return false, errors.Wrap(err, "Get -> StartService")
	}
	res_ := res.Raw().Value().(int32)
	fmt.Println(res_)
	if res_ == 0 || res_ == 10 {
		return true, nil

	}
	return false, nil
}

// We can not use the IsProcessorFeaturePresent approach, as it does not matter in self-virtualized environment
// see https://devblogs.microsoft.com/oldnewthing/20201216-00/?p=104550
func (m *Manager) HasVTx() (bool, error) {
	con, _ := wmiAPIInit()
	defer wmiAPIRelease(con)

	res, err := con.CallMethod("ExecQuery", "SELECT * FROM Win32_ComputerSystem")
	if err != nil {
		return false, errors.Wrap(err, "ExecQuery")
	}
	item, err := res.ItemAtIndex(0)
	if err != nil {
		return false, errors.Wrap(err, "ItemAtIndex")
	}
	prop_, err := item.GetProperty("HypervisorPresent")
	if err != nil {
		return false, errors.Wrap(err, "GetProperty")
	}
	return prop_.Value().(bool), nil
}

func (m *Manager) EnableHyperVPlatform() error {
	log.Println("EnableHyperVPlatform > May take ~5 min.")

	// // necessary for Home Edition
	// packagesPath := os.Getenv("SYSTEMROOT") + `\servicing\Packages\`
	// err := filepath.Walk(packagesPath, func(path string, info fs.FileInfo, _ error) error {
	// 	//log.Println("info>", info)
	// 	if info.IsDir() {
	// 		return nil
	// 	}
	// 	if strings.HasSuffix(info.Name(), ".mum") && strings.Contains(info.Name(), "Hyper-V") {
	// 		p := packagesPath + info.Name()

	// 		exe := "dism.exe"
	// 		cmdArgs := fmt.Sprintf("/online /norestart /add-package:%s", p)
	// 		err := native.ShellExecuteAndWait(0, "runas", exe, cmdArgs, "", syscall.SW_HIDE)
	// 		if err != nil {
	// 			log.Println("Command failed: failed to enable " + p)
	// 			return err
	// 		}
	// 	}
	// 	return nil
	// })
	// if err != nil {
	// 	return err
	// }
	var err error

	if !m.hasDism {
		m.initializeDism()
	}
	for _, f := range features {
		log.Println("EnableFeature>", f)

		err = m.ses.EnableFeature(f, "", nil, true, nil, nil)
		if err != nil {
			success := errors.Is(err, windows.ERROR_SUCCESS_REBOOT_REQUIRED) || errors.Is(err, windows.ERROR_SUCCESS_RESTART_REQUIRED)
			log.Println("err>", err, success)

			if !success {
				return err
			}
		}
	}

	return nil
}

func (m *Manager) initializeDism() error {
	ses, err := dism.OpenSession(dism.DISM_ONLINE_IMAGE, "", "", dism.DismLogErrorsWarningsInfo, "", "")
	if err != nil {
		return err
	}
	m.ses = ses
	m.hasDism = true
	return nil
}
