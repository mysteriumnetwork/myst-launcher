/**
 * Copyright (c) 2021 BlockDev AG
 *
 * This source code is licensed under the MIT license found in the
 * LICENSE file in the root directory of this source tree.
 */

package gui_win32

import "github.com/lxn/walk"

func ErrorModal(title, message string) int {
	return walk.MsgBox(nil, title, message, walk.MsgBoxTopMost|walk.MsgBoxOK|walk.MsgBoxIconError)
}
