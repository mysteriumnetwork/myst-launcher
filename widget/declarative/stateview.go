//go:build windows
// +build windows

/**
 * Copyright (c) 2021 BlockDev AG
 *
 * This source code is licensed under the MIT license found in the
 * LICENSE file in the root directory of this source tree.
 */

package declarative

import (
	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"

	"github.com/mysteriumnetwork/myst-launcher/widget/impl"
)

type StatusView struct {
	// Window

	Accessibility      Accessibility
	Background         Brush
	ContextMenuItems   []MenuItem
	DoubleBuffering    bool
	Enabled            Property
	Font               Font
	MaxSize            Size
	MinSize            Size
	Name               string
	OnBoundsChanged    walk.EventHandler
	OnKeyDown          walk.KeyEventHandler
	OnKeyPress         walk.KeyEventHandler
	OnKeyUp            walk.KeyEventHandler
	OnMouseDown        walk.MouseEventHandler
	OnMouseMove        walk.MouseEventHandler
	OnMouseUp          walk.MouseEventHandler
	OnSizeChanged      walk.EventHandler
	Persistent         bool
	RightToLeftReading bool
	ToolTipText        Property
	Visible            Property

	// Widget

	Alignment          Alignment2D
	AlwaysConsumeSpace bool
	Column             int
	ColumnSpan         int
	GraphicsEffects    []walk.WidgetGraphicsEffect
	Row                int
	RowSpan            int
	StretchFactor      int

	// CustomWidget

	AssignTo            **impl.StatusViewImpl
	ClearsBackground    bool
	InvalidatesOnResize bool
	PaintMode           PaintMode
	Style               uint32
}

func (cw StatusView) Create(builder *Builder) error {

	w, err := impl.NewCustomWidget2Pixels(builder.Parent(), uint(cw.Style))
	if err != nil {
		return err
	}

	if cw.AssignTo != nil {
		*cw.AssignTo = w
	}

	return builder.InitWidget(cw, w, func() error {

		w.SetClearsBackground(cw.ClearsBackground)
		w.SetInvalidatesOnResize(cw.InvalidatesOnResize)
		w.SetPaintMode(walk.PaintMode(cw.PaintMode))

		return nil
	})
}
