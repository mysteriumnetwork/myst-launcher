/**
 * Copyright (c) 2021 BlockDev AG
 *
 * This source code is licensed under the MIT license found in the
 * LICENSE file in the root directory of this source tree.
 */

package _const

var ImageTag = "latest"

const (
	ImageName = "mysteriumnetwork/myst"
)

func GetImageName() string {
	return ImageName + ":" + ImageTag
}
