//go:build darwin
// +build darwin

/**
 * Copyright (c) 2021 BlockDev AG
 *
 * This source code is licensed under the MIT license found in the
 * LICENSE file in the root directory of this source tree.
 */

package utils

import (
	"bytes"
	"log"
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
	log.Println("QueryFeatures >")
	f := make([]int, 0)

	hasHV := func() bool {
		var out bytes.Buffer
		_, err := CmdRun(&out, "sysctl", "machdep.cpu.features")
		if err != nil {
			log.Println("QueryFeatures >", err)
			return false
		}
		if !strings.Contains(out.String(), "VMX") {
			return false
		}
		out.Reset()

		_, err = CmdRun(&out, "sysctl", "kern.hv_support")
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
		f = append(f, IDFeatureHyperisorFramework)
	}
	return f, nil
}

func InstallFeatures(features []int, onFeatureReady func(int, string)) error {
	return nil
}
