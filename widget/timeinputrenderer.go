package widget

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

//- timeInputRenderer

type timeInputRenderer struct {
	input                       *TimeInput
	boxAM, line                 *canvas.Rectangle
	bgHour, bgMin, bgSec        *canvas.Rectangle
	cursor                      *canvas.Rectangle
	textHour, textMin, textSec  *canvas.Text
	separatorHour, separatorMin *canvas.Text
	dayPart                     *canvas.Text
	wheel                       *mouseWheel
	boxMinSize                  fyne.Size
	separatorMinSize            fyne.Size
}

// Destroy is for internal use.
func (r *timeInputRenderer) Destroy() {
	r.input.cursorAnim.Stop()
}

// MinSize returns the minimum size of the widget that is rendered by this renderer.
func (r *timeInputRenderer) MinSize() fyne.Size {
	// minimal with only hour & minute
	w := theme.Padding()*4 + r.boxMinSize.Width*2 + r.separatorMinSize.Width
	if r.input.showSeconds {
		w += r.boxMinSize.Width + r.separatorMinSize.Width
	}
	if r.input.clock12Hour {
		w += r.separatorMinSize.Width + r.boxMinSize.Width
	}
	h := theme.Padding()*4 + r.boxMinSize.Height
	return fyne.NewSize(w, h)
}

// Layout is a hook that is called if the widget needs to be laid out.
// This should never call Refresh.
func (r *timeInputRenderer) Layout(size fyne.Size) {
	pad := theme.Padding()
	pos := fyne.NewPos(pad*2, pad*2)

	r.wheel.Move(fyne.Position{})
	r.wheel.Resize(size)

	r.line.Move(fyne.NewPos(0, size.Height-pad/2))
	r.line.Resize(fyne.NewSize(size.Width, pad/2))

	r.bgHour.Move(pos)
	r.bgHour.Resize(r.boxMinSize)
	r.textHour.Move(pos)
	r.textHour.Resize(r.boxMinSize)
	r.input.sections[0].pos = pos
	r.input.sections[0].size = r.boxMinSize
	pos = pos.Add(fyne.NewDelta(r.boxMinSize.Width, 0))

	r.separatorHour.Move(pos)
	r.separatorHour.Resize(r.separatorMinSize)
	pos = pos.Add(fyne.NewDelta(r.separatorMinSize.Width, 0))

	r.textMin.Move(pos)
	r.bgMin.Move(pos)
	r.textMin.Resize(r.boxMinSize)
	r.bgMin.Resize(r.boxMinSize)
	r.input.sections[1].pos = pos
	r.input.sections[1].size = r.boxMinSize
	pos = pos.Add(fyne.NewDelta(r.boxMinSize.Width, 0))

	idx := 2
	if r.input.showSeconds {
		r.separatorMin.Move(pos)
		r.separatorMin.Resize(r.separatorMinSize)
		pos = pos.Add(fyne.NewDelta(r.separatorMinSize.Width, 0))

		r.textSec.Move(pos)
		r.bgSec.Move(pos)
		r.textSec.Resize(r.boxMinSize)
		r.bgSec.Resize(r.boxMinSize)
		r.input.sections[idx].pos = pos
		r.input.sections[idx].size = r.boxMinSize
		pos = pos.Add(fyne.NewDelta(r.boxMinSize.Width+r.separatorMinSize.Width, 0))
		idx++
	} else {
		pos = pos.Add(fyne.NewDelta(r.separatorMinSize.Width, 0))
	}

	if r.input.clock12Hour {
		pos = pos.Subtract(fyne.NewDelta(pad, 0))
		r.dayPart.Move(pos)
		r.boxAM.Move(pos)
		r.dayPart.Resize(r.boxMinSize)
		r.boxAM.Resize(r.boxMinSize)
		r.input.sections[idx].pos = pos
		r.input.sections[idx].size = r.boxMinSize
	}
}

// Objects returns all objects that should be drawn.
func (r *timeInputRenderer) Objects() []fyne.CanvasObject {
	objects := []fyne.CanvasObject{
		r.wheel, r.line, r.bgHour, r.bgMin,
		r.textHour,
		r.separatorHour,
		r.textMin,
	}
	if r.input.showSeconds {
		objects = append(objects, r.bgSec, r.separatorMin, r.textSec)
	}
	if r.input.clock12Hour {
		objects = append(objects, r.boxAM, r.dayPart)
	}
	objects = append(objects, r.input.cursorAnim.Cursor)
	return objects
}

// Refresh is a hook that is called if the widget has updated and needs to be redrawn.
// This might trigger a Layout.
func (r *timeInputRenderer) Refresh() {
	r.makeSureOfComponents()

	var lineColor color.Color
	if r.input.showFocus {
		lineColor = theme.PrimaryColor()
	} else {
		lineColor = theme.ShadowColor()
	}
	setRectangleColor(r.line, lineColor)

	var bgColor color.Color
	// the hours are selected
	setRectangleColor(r.bgHour, r.colorForSection(0))
	setCancasText(r.textHour, r.input.sections[0].get())

	setRectangleColor(r.bgMin, r.colorForSection(1))
	setCancasText(r.textMin, r.input.sections[1].get())

	idx := 2
	if r.input.showSeconds {
		setRectangleColor(r.bgSec, r.colorForSection(idx))
		setCancasText(r.textSec, r.input.sections[idx].get())
		idx++
	}

	if r.input.clock12Hour {
		if r.input.showFocus && r.input.section == idx {
			bgColor = theme.PrimaryColor()
		} else {
			bgColor = theme.InputBackgroundColor()
		}
		setRectangleColor(r.boxAM, bgColor)
		setCancasText(r.dayPart, r.input.sections[idx].get())
	}
	if r.input.showFocus {
		r.cursor.Show()
		r.input.cursorAnim.Start()
		r.moveCursor()
	} else {
		r.input.cursorAnim.Stop()
		r.cursor.Hide()
	}
}

func setCancasText(r *canvas.Text, s string) {
	if r.Text != s {
		r.Text = s
		r.Refresh()
	}
}

func setRectangleColor(r *canvas.Rectangle, color color.Color) {
	if r.FillColor != color {
		r.FillColor = color
		r.Refresh()
	}
}

func (r *timeInputRenderer) colorForSection(section int) color.Color {
	if !r.input.sections[section].valid() {
		return theme.ErrorColor()
	}
	if r.input.showFocus && r.input.selection == section {
		return theme.PrimaryColor()
	}
	return theme.InputBackgroundColor()
}

func (r *timeInputRenderer) lineSizeToColumn(section int) fyne.Size {
	minSize := fyne.NewSize(-1, -1)
	switch section {
	case 0:
		minSize = r.textHour.MinSize()
	case 1:
		minSize = r.textMin.MinSize()
	case 2:
		if r.input.showSeconds {
			minSize = r.textSec.MinSize()
		}
	}
	return minSize
}

func (r *timeInputRenderer) makeSureOfComponents() {
	if r.input.showSeconds {
		if r.textSec == nil {
			r.bgSec = canvas.NewRectangle(theme.InputBackgroundColor())
			r.textSec = &canvas.Text{Text: "00", Alignment: fyne.TextAlignCenter}
			r.separatorMin = &canvas.Text{Text: ":", Alignment: fyne.TextAlignCenter, TextStyle: fyne.TextStyle{Monospace: true}}
		}
	} else {
		r.bgSec = nil
		r.textSec = nil
		r.separatorMin = nil
	}
	if r.input.clock12Hour {
		if r.dayPart == nil {
			r.boxAM = canvas.NewRectangle(theme.ShadowColor())
			r.dayPart = &canvas.Text{Text: "AM", Alignment: fyne.TextAlignCenter}
			r.boxMinSize = r.dayPart.MinSize().Add(fyne.NewDelta(theme.Padding()*2, theme.Padding())).Max(r.boxMinSize)
		}
	} else {
		r.boxAM = nil
		r.dayPart = nil
	}
}

func (r *timeInputRenderer) moveCursor() {
	xPos := float32(0)
	if r.input.selection >= 0 {
		// cursor at the end of selection box
		xPos = r.boxMinSize.Width
		if r.input.clock12Hour && r.input.section == len(r.input.sections)-1 {
			xPos -= theme.Padding()
		}
	} else {
		inputSize := r.lineSizeToColumn(r.input.section)
		if inputSize.Width < 0 {
			// hide cursor
			r.cursor.Hide()
			return
		}
		xPos = inputSize.Width + (r.boxMinSize.Width-inputSize.Width)/2
	}
	xPos += float32(r.input.section) * (r.boxMinSize.Width + r.separatorMinSize.Width)

	if r.cursor.Hidden {
		r.cursor.Show()
	}

	lineHeight := r.boxMinSize.Height
	r.cursor.Resize(fyne.NewSize(2, lineHeight))
	r.cursor.Move(fyne.NewPos(xPos-1+theme.Padding()*2, theme.Padding()*2))
}

var _ fyne.Scrollable = (*mouseWheel)(nil)

// mouseWheel receiver for mouse scroll events
type mouseWheel struct {
	widget.BaseWidget
	OnScroll func(*fyne.ScrollEvent)
}

func (w *mouseWheel) Scrolled(e *fyne.ScrollEvent) {
	if w.OnScroll != nil {
		w.OnScroll(e)
	}
}
