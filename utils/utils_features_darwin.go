// +build darwin

package utils

import (
	"bytes"
	"fmt"
	"strings"
)

const (
	FeatureHyperisorFramework = "HyperisorFramework"

	IDFeatureHyperisorFramework = 1
)

var featureDict = map[int]string{
	IDFeatureHyperisorFramework: FeatureHyperisorFramework,
}

// query if there are features to be enabled
func QueryFeatures() ([]int, error) {
	fmt.Println("QueryFeatures >")
	f := make([]int, 0)

	hasHV := func() bool {
		var out bytes.Buffer
		_, err := CmdRun(&out, "sysctl", "machdep.cpu.features")
		if err != nil {
			fmt.Println("QueryFeatures >", err)
			return false
		}
		fmt.Println("QueryFeatures >", out.String())
		if !strings.Contains(out.String(), "VMX") {
			return false
		}
		out.Reset()

		_, err = CmdRun(&out, "sysctl", "kern.hv_support")
		if err != nil {
			fmt.Println("QueryFeatures >", err)
			return false
		}
		fmt.Println("QueryFeatures >", out.String())
		if !strings.HasPrefix(out.String(), "kern.hv_support: 1") {
			return false
		}
		//return true
		return false
	}()
	if !hasHV {
		f = append(f, IDFeatureHyperisorFramework)
	}
	return f, nil
}

func InstallFeatures(features []int, onFeatureReady func(int, string)) error {
	// for _, feature := range features {
	// 	featureName := featureDict[feature]

	// 	log.Println("Enable " + featureName)
	// 	exe := "dism.exe"
	// 	cmdArgs := fmt.Sprintf("/online /enable-feature /featurename:%s /all /norestart", featureName)
	// 	err := native.ShellExecuteAndWait(0, "runas", exe, cmdArgs, "", syscall.SW_HIDE)
	// 	if err != nil {
	// 		log.Println("Command failed: failed to enable" + featureName)
	// 		return err
	// 	}

	// 	if onFeatureReady != nil {
	// 		onFeatureReady(feature, featureName)
	// 	}
	// }
	return nil
}
