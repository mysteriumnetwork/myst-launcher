// +build !windows

package utils

import "os"

func GetTmpDir() string {
	res := os.Getenv("TEMPDIR")
	if res == "" {
		res = "/tmp"
	}
	return res
}

func GetUserProfileDir() string {
	return os.Getenv("HOME")
}
