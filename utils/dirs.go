// +build !windows

package utils

import "os"

func GetTmpDir() string {
	return "/tmp"
}

func GetUserProfileDir() string {
	return os.Getenv("HOME")
}
