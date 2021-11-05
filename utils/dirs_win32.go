//go:build windows
// +build windows

/**
 * Copyright (c) 2021 BlockDev AG
 *
 * This source code is licensed under the MIT license found in the
 * LICENSE file in the root directory of this source tree.
 */

package utils

import "os"

func GetTmpDir() string {
	return os.Getenv("TMP")
}

func GetUserProfileDir() string {
	return os.Getenv("USERPROFILE")
}
