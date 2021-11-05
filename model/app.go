/**
 * Copyright (c) 2021 BlockDev AG
 *
 * This source code is licensed under the MIT license found in the
 * LICENSE file in the root directory of this source tree.
 */

package model

type AppInterface interface {
	TriggerAction(action string)
	GetInTray() bool
}
