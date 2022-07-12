//go:build windows
// +build windows

package native

import (
	"github.com/artdarek/go-unzip"
)

func extractNodeBinary(src, dest string) error {
	return unzip.New(src, dest).Extract()
}
