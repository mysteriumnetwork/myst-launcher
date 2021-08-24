// +build windows

package utils

import "os"

func GetTmpDir() string {
	return os.Getenv("TMP")
}

func GetUserProfileDir() string {
	return os.Getenv("USERPROFILE")
}
