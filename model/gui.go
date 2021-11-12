/**
 * Copyright (c) 2021 BlockDev AG
 *
 * This source code is licensed under the MIT license found in the
 * LICENSE file in the root directory of this source tree.
 */

package model

type Gui_ interface {
	CloseUI()

	OpenDialogue(id int)
	DialogueComplete()
	WaitDialogueComplete() bool
	TerminateWaitDialogueComplete()

	PopupMain()
	ShowMain()
	ShowNotificationInstalled()
	ShowNotificationUpgrade()

	ConfirmModal(title, message string) int
	YesNoModal(title, message string) int
	ErrorModal(title, message string) int
	SetModalReturnCode(rc int)
}
