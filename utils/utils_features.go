//go:build windows
// +build windows

/**
 * Copyright (c) 2021 BlockDev AG
 *
 * This source code is licensed under the MIT license found in the
 * LICENSE file in the root directory of this source tree.
 */

package utils

import (
	"fmt"
	"log"
	"syscall"

	"github.com/mysteriumnetwork/myst-launcher/native"
)

const (
	FeatureWSL        = "Microsoft-Windows-Subsystem-Linux"
	FeatureHyperV     = "Microsoft-Hyper-V"
	FeatureVMPlatform = "VirtualMachinePlatform"

	IDFeatureWSL        = 1
	IDFeatureHyperV     = 2
	IDFeatureVMPlatform = 3
)

var featureDict = map[int]string{
	IDFeatureWSL:        FeatureWSL,
	IDFeatureHyperV:     FeatureHyperV,
	IDFeatureVMPlatform: FeatureVMPlatform,
}

// query if there are features to be enabled
func QueryFeatures() ([]int, error) {
	f := make([]int, 0)
	for k := 1; k <= 3; k++ {
		v := featureDict[k]
		featureExists, featureEnabled, err := QueryWindowsFeature(v)
		if err != nil {
			return nil, err
		}

		if featureExists && !featureEnabled {
			f = append(f, k)
		}
	}
	return f, nil
}

func InstallFeatures(features []int, onFeatureReady func(int, string)) error {
	for _, feature := range features {
		featureName := featureDict[feature]

		log.Println("Enable " + featureName)
		exe := "dism.exe"
		cmdArgs := fmt.Sprintf("/online /enable-feature /featurename:%s /all /norestart", featureName)
		err := native.ShellExecuteAndWait(0, "runas", exe, cmdArgs, "", syscall.SW_HIDE)
		if err != nil {
			log.Println("Command failed: failed to enable" + featureName)
			return err
		}
		if onFeatureReady != nil {
			onFeatureReady(feature, featureName)
		}
	}
	return nil
}
