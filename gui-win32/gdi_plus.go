/**
 * Copyright (c) 2021 BlockDev AG
 *
 * This source code is licensed under the MIT license found in the
 * LICENSE file in the root directory of this source tree.
 */

package gui_win32

import (
	"github.com/tryor/gdiplus"
	"github.com/tryor/winapi"
	"log"
)

var (
	gpToken winapi.ULONG_PTR
	input   gdiplus.GdiplusStartupInput
)

func InitGDIPlus() {
	log.Println("Initializing GDI+ ....")
	input.GdiplusVersion = 1
	_, err := gdiplus.Startup(&gpToken, &input, nil)
	if err != nil {
		panic(err)
	}
}

func ShutdownGDIPlus() {
	gdiplus.GdiplusShutdown(gpToken)
}
