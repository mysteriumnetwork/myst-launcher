package app

const (
	FeatureWSL        = "Microsoft-Windows-Subsystem-Linux"
	FeatureHyperV     = "Microsoft-Hyper-V"
	FeatureVMPlatform = "VirtualMachinePlatform"

	FeatureWSL_        = 1
	FeatureHyperV_     = 2
	FeatureVMPlatform_ = 3
)

var fm = map[int]string{
	FeatureWSL_:        FeatureWSL,
	FeatureHyperV_:     FeatureHyperV,
	FeatureVMPlatform_: FeatureVMPlatform,
}

// query if there are features to be enabled
func QueryFeatures() ([]int, error) {
	f := make([]int, 0)
	for k, v := range fm {
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
