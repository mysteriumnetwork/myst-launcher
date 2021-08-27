// +build !windows

package utils

import "os"

func GetTmpDir() string {
	return os.Getenv("TEMPDIR")
}

func GetUserProfileDir() string {
	return os.Getenv("HOME")
}
