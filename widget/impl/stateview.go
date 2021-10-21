// Copyright 2010 The Walk Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build windows
// +build windows

package impl

import (
	"errors"
	"fmt"
	"log"
	"unsafe"

	"github.com/gonutz/w32"
	"github.com/lxn/walk"
	"github.com/lxn/win"
	"github.com/tryor/gdiplus"
	"github.com/tryor/winapi"

	"github.com/mysteriumnetwork/myst-launcher/native/resource"
)

const statusViewWindowClass = `\o/ Walk_StatusWidget_Class \o/`

func init() {
	walk.AppendToWalkInit(func() {
		walk.MustRegisterWindowClass(statusViewWindowClass)
	})
}

var (
	red   = gdiplus.NewColor2(255, 59, 48)
	green = gdiplus.NewColor2(52, 199, 89)
)

type StatusViewImpl struct {
	walk.WidgetBase

	// paint               PaintFunc2 // in 1/96" units
	// paintPixels         PaintFunc // in native pixels
	invalidatesOnResize bool
	paintMode           walk.PaintMode

	// 0 - off
	// 1 - red
	// 2 - green
	// 3 - spinner
	state int

	im            *gdiplus.Image
	gifFrame      int
	gifFrameCount int
}

// NewCustomWidget2Pixels creates and initializes a new custom draw widget.
func NewCustomWidget2Pixels(parent walk.Container, style uint) (*StatusViewImpl, error) {
	cw := &StatusViewImpl{}
	err := cw.init(parent, style)
	if err != nil {
		return nil, err
	}
	return cw, nil
}

var imgCache map[string]*gdiplus.Image

func init() {
	imgCache = make(map[string]*gdiplus.Image, 0)
}

func (cw *StatusViewImpl) init(parent walk.Container, style uint) error {
	if err := walk.InitWidget(
		cw,
		parent,
		statusViewWindowClass,
		win.WS_VISIBLE|uint32(style),
		0); err != nil {
		return err
	}

	resName := "spinner-18px"
	im, ok := imgCache[resName]
	if !ok {
		var err error
		hh, err := resource.FindByName(0, "SPINNER", resource.RT_RCDATA)
		if err != nil {
			log.Println(err)
			return err
		}
		d, err := resource.Load(0, hh)
		if err != nil {
			return err
		}
		iStream := w32.SHCreateMemStream(d)
		_ = iStream
		im, err = gdiplus.NewImageFromStream((*winapi.IStream)(unsafe.Pointer(iStream)))
		if err != nil {
			log.Println(err)
			return err
		}
		imgCache[resName] = im
	}

	count := im.GetFrameDimensionsCount()
	dimensionIDs, _ := im.GetFrameDimensionsList(count)
	cw.gifFrameCount = int(im.GetFrameCount(&dimensionIDs[0]))
	cw.im = im

	return nil
}

// deprecated, use PaintMode
func (cw *StatusViewImpl) ClearsBackground() bool {
	return cw.paintMode != walk.PaintNormal
}

// deprecated, use SetPaintMode
func (cw *StatusViewImpl) SetClearsBackground(value bool) {
	if value != cw.ClearsBackground() {
		if value {
			cw.paintMode = walk.PaintNormal
		} else {
			cw.paintMode = walk.PaintNoErase
		}
	}
}

func (cw *StatusViewImpl) InvalidatesOnResize() bool {
	return cw.invalidatesOnResize
}

func (cw *StatusViewImpl) SetInvalidatesOnResize(value bool) {
	cw.invalidatesOnResize = value
}

func (cw *StatusViewImpl) PaintMode() walk.PaintMode {
	return cw.paintMode
}

func (cw *StatusViewImpl) SetPaintMode(value walk.PaintMode) {
	cw.paintMode = value
}

func (cw *StatusViewImpl) SetState(value int) {
	cw.state = value

	switch value {
	case 3:
		win.SetTimer(cw.Handle(), 2000, 100, 0)
	case 0, 1, 2:
		win.KillTimer(cw.Handle(), 2000)
	}

	cw.Invalidate()
	win.UpdateWindow(cw.Handle())
}

func (cw *StatusViewImpl) paint(hdc win.HDC, updateBounds walk.Rectangle, state int) error {

	gp, err := gdiplus.NewGraphicsFromHDC(winapi.HDC(hdc))
	if err != nil {
		fmt.Println(err)
	}

	switch cw.state {
	case 3:
		gp.DrawImage3(cw.im, 2, 0, gdiplus.REAL(cw.im.GetWidth()), gdiplus.REAL(cw.im.GetHeight()))
	case 0:

	case 1, 2:
		c := red
		if state == 2 {
			c = green
		}
		br, err := gdiplus.NewSolidBrush(c)
		if err != nil {
			fmt.Println(err)
		}
		gp.SetTextRenderingHint(gdiplus.TextRenderingHintAntiAlias)
		gp.SetSmoothingMode(gdiplus.SmoothingModeAntiAlias)
		gp.FillEllipseI(br, 4, 4, 12, 12)
	}
	gp.Release()
	return nil
}

func newError(message string) error {
	return errors.New(message)
}

func rectangleFromRECT(r win.RECT) walk.Rectangle {
	return walk.Rectangle{
		X:      int(r.Left),
		Y:      int(r.Top),
		Width:  int(r.Right - r.Left),
		Height: int(r.Bottom - r.Top),
	}
}

func (cw *StatusViewImpl) paintProc(wParam, lParam uintptr) error {
	var ps win.PAINTSTRUCT
	var hdc win.HDC
	if wParam == 0 {
		hdc = win.BeginPaint(cw.Handle(), &ps)
	} else {
		hdc = win.HDC(wParam)
	}
	if hdc == 0 {
		return newError("BeginPaint failed")
	}
	defer func() {
		if wParam == 0 {
			win.EndPaint(cw.Handle(), &ps)
		}
	}()

	bounds := rectangleFromRECT(ps.RcPaint)
	return cw.paint(hdc, walk.RectangleTo96DPI(bounds, cw.DPI()), cw.state)
}

func (cw *StatusViewImpl) WndProc(hwnd win.HWND, msg uint32, wParam, lParam uintptr) uintptr {
	switch msg {
	case win.WM_TIMER:
		if wParam == 2000 {
			cw.im.SelectActiveFrame(gdiplus.FrameDimensionTime, winapi.UINT(cw.gifFrame))
			cw.gifFrame++
			if cw.gifFrame == cw.gifFrameCount {
				cw.gifFrame = 0
			}
			win.InvalidateRect(hwnd, nil, false)
		}

	case win.WM_PAINT:
		err := cw.paintProc(wParam, lParam)
		if err != nil {
			newError("paint failed")
			break
		}
		return 0

	case win.WM_ERASEBKGND:
		if cw.paintMode != walk.PaintNormal {
			return 1
		}

	case win.WM_PRINTCLIENT:
		win.SendMessage(hwnd, win.WM_PAINT, wParam, lParam)

	case win.WM_WINDOWPOSCHANGED:
		wp := (*win.WINDOWPOS)(unsafe.Pointer(lParam))

		if wp.Flags&win.SWP_NOSIZE != 0 {
			break
		}

		if cw.invalidatesOnResize {
			cw.Invalidate()
		}
	}

	return cw.WidgetBase.WndProc(hwnd, msg, wParam, lParam)
}

func (*StatusViewImpl) CreateLayoutItem(ctx *walk.LayoutContext) walk.LayoutItem {
	// return &myWidgetLayoutItem{idealSize: walk.SizeFrom96DPI(walk.Size{20, 20}, ctx.DPI())}
	return &myWidgetLayoutItem{idealSize: walk.Size{20, 20}}
}

type myWidgetLayoutItem struct {
	walk.LayoutItemBase
	idealSize walk.Size // in native pixels
}

func (li *myWidgetLayoutItem) LayoutFlags() walk.LayoutFlags {
	return 0
}

func (li *myWidgetLayoutItem) IdealSize() walk.Size {
	return li.idealSize
}
