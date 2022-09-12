//go:build !windows
// +build !windows

package native

import (
	"code.cloudfoundry.org/archiver/extractor"
)

func extractNodeBinary(src, dest string) error {
	return extractor.NewTgz().Extract(src, dest)
}

func CheckAndInstallFirewall() {
}