/*
timeinput provide time input component.
The best way to create new widget is to use NewTimeInput function.
The widget initialized either to use or not seconds, and

*/
package widget

import (
	"context"
	"image/color"
	"sync"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// Clock12Hour tell to use 12 or 24 clock. Default is 12 hour clock.
var Clock12Hour bool = true

const doubleClickDelay = 300 // ms (maximum interval between clicks for double click detection)

func init() {
	// TODO: find how to properly set Clock12Hour
}

// NewTimeInput return new TimeInput widget
func NewTimeInput(showSeconds bool) *TimeInput {
	t := &TimeInput{showSeconds: showSeconds, clock12Hour: Clock12Hour}
	t.makeSureOfSections()
	t.clock = time.Now()
	t.ExtendBaseWidget(t)
	return t
}

type TimeInput struct {
	widget.DisableableWidget

	clock         time.Time
	clock12Hour   bool
	section       int
	selection     int
	sections      []section
	numericFields int
	showSeconds   bool
	showFocus     bool
	cursorAnim    *EntryCursorAnimation
	shortcut      fyne.ShortcutHandler
	propertyLock  sync.Mutex
	tapLock       sync.Mutex
	tapLast       int
	tapCancelFunc func()
}

func (t *TimeInput) Set(v time.Time) {
	t.makeSureOfSections()

	t.propertyLock.Lock()
	if t.showSeconds {
		t.clock = v.Round(time.Second)
	} else {
		t.clock = v.Round(time.Minute)
	}
	t.updateSections()
	t.selection = 0
	t.section = 0
	t.sections[t.section].col = t.sections[t.section].maxCol
	t.propertyLock.Unlock()

	t.Refresh()
}

func (t *TimeInput) Get() time.Time {
	t.propertyLock.Lock()
	defer t.propertyLock.Unlock()
	if t.valid() {
		year, month, day := t.clock.Date()
		h := t.sections[0].value
		s := 0
		idx := 2
		if t.showSeconds {
			s = t.sections[idx].value
			idx++
		}
		if t.clock12Hour {
			if t.sections[idx].value == 1 {
				h += 12
			}
		}
		t.clock = time.Date(year, month, day, h, t.sections[1].value, s, 0, t.clock.Location())
	}
	return t.clock
}

// Confirm to interfaces:
var _ fyne.Widget = (*TimeInput)(nil)
var _ fyne.Focusable = (*TimeInput)(nil)
var _ fyne.Tappable = (*TimeInput)(nil)

func (t *TimeInput) CreateRenderer() fyne.WidgetRenderer {
	t.makeSureOfSections()

	t.propertyLock.Lock()
	defer t.propertyLock.Unlock()

	t.updateSections()
	t.selection = -1
	t.section = 0

	cursor := canvas.NewRectangle(color.Transparent)
	cursor.Hide()
	t.cursorAnim = NewEntryCursorAnimation(cursor)

	r := &timeInputRenderer{input: t}
	r.cursor = t.cursorAnim.Cursor
	r.wheel = &mouseWheel{}
	r.wheel.OnScroll = t.onWheelScroll
	r.line = canvas.NewRectangle(theme.ShadowColor())
	if t.showFocus {
		r.line.FillColor = theme.PrimaryColor()
	}
	r.textHour = &canvas.Text{Text: "00", TextSize: theme.TextSize(), Alignment: fyne.TextAlignCenter}
	r.bgHour = canvas.NewRectangle(theme.InputBackgroundColor())
	r.separatorHour = &canvas.Text{Text: ":", Alignment: fyne.TextAlignCenter, TextStyle: fyne.TextStyle{Monospace: true}, TextSize: theme.TextSize()}
	r.textMin = &canvas.Text{Text: "00", TextSize: theme.TextSize(), Alignment: fyne.TextAlignCenter}
	r.bgMin = canvas.NewRectangle(theme.InputBackgroundColor())

	r.boxMinSize = r.textHour.MinSize().Add(fyne.NewDelta(theme.Padding()*2, theme.Padding()))
	r.separatorMinSize = r.separatorHour.MinSize().Add(fyne.NewDelta(0, theme.Padding()))

	r.makeSureOfComponents()
	return r
}

// FocusGained is a hook called by the focus handling logic after this object gained the focus.
func (t *TimeInput) FocusGained() {
	t.propertyLock.Lock()
	if !t.showFocus {
		t.showFocus = true
		t.selection = 0
		t.section = 0
		t.sections[t.section].col = t.sections[t.section].maxCol
		defer t.Refresh()
	}
	t.propertyLock.Unlock()
}

// FocusLost is a hook called by the focus handling logic after this object lost the focus.
func (t *TimeInput) FocusLost() {
	t.propertyLock.Lock()
	if t.showFocus {
		t.showFocus = false
		defer t.Refresh()
	}
	t.propertyLock.Unlock()
}

func (t *TimeInput) Tapped(p *fyne.PointEvent) {
	t.tapLock.Lock()
	t.propertyLock.Lock()
	s := t.getSection(p.Position)
	if s >= 0 {
		t.section = s
		t.selection = s
		t.sections[t.section].col = t.sections[t.section].maxCol
		if t.clock12Hour && s == t.numericFields {
			if t.tapLast == s {
				t.sections[s].runePressed(' ')
			} else {
				t.tapLast = s
				go t.startDoubleTapTimer()
			}
		}
		defer t.Refresh()
	}
	t.propertyLock.Unlock()
	t.tapLock.Unlock()
}

// TypedKey is a hook called by the input handling logic on key events if this object is focused.
func (t *TimeInput) TypedKey(key *fyne.KeyEvent) {
	if t.Disabled() {
		return
	}
	if t.cursorAnim != nil {
		t.cursorAnim.Interrupt()
	}

	switch key.Name {
	case fyne.KeyReturn, fyne.KeyEnter:
		defer fyne.CurrentApp().Driver().CanvasForObject(t).FocusNext()

	case fyne.KeyLeft, fyne.KeyRight, fyne.KeyUp, fyne.KeyDown:
		if t.sections[t.section].keyPressed(key) {
			defer t.Refresh()
		}

	case fyne.KeyEnd:
		t.propertyLock.Lock()
		t.selection = -1
		t.section = len(t.sections) - 1
		t.sections[t.section].col = t.sections[t.section].maxCol
		defer t.Refresh()
		t.propertyLock.Unlock()

	case fyne.KeyHome:
		t.propertyLock.Lock()
		t.selection = 0
		t.section = 0
		t.sections[t.section].col = t.sections[t.section].maxCol
		defer t.Refresh()
		t.propertyLock.Unlock()
	}
}

// TypedRune is a hook called by the input handling logic on text input events if this object is focused.
func (t *TimeInput) TypedRune(r rune) {
	if t.Disabled() {
		return
	}
	t.propertyLock.Lock()
	t.unSelect(true)
	if t.sections[t.section].runePressed(r) {
		defer t.Refresh()
	}
	t.propertyLock.Unlock()
}

// TypedShortcut implements the Shortcutable interface
//
// Implements: fyne.Shortcutable
func (t *TimeInput) TypedShortcut(shortcut fyne.Shortcut) {
	t.shortcut.TypedShortcut(shortcut)
}

// copyToClipboard copies the current selection to a given clipboard.
// This does nothing if it is a concealed entry.
func (t *TimeInput) copyToClipboard(clipboard fyne.Clipboard) {
	clipboard.SetContent(t.clock.Local().Format("15:04:05"))
}

func (t *TimeInput) getSection(p fyne.Position) int {
	for idx, s := range t.sections {
		a := p.Subtract(s.pos)
		if a.X > 0 && a.Y > 0 {
			a = s.pos.Add(s.size).Subtract(p)
			if a.X > 0 && a.Y > 0 {
				return idx
			}
		}
	}
	return -1
}

func (t *TimeInput) makeSureOfSections() {
	t.propertyLock.Lock()
	defer t.propertyLock.Unlock()

	sections := 2
	if t.showSeconds {
		sections++
	}
	t.numericFields = sections
	if t.clock12Hour {
		sections++
	}

	if t.sections == nil {
		t.registerShortcut()
	}

	if t.sections == nil || len(t.sections) != sections {
		t.sections = make([]section, sections)

		t.sections[0].maxCol = 2
		t.sections[0].maxValue = 23
		t.sections[0].onNextSection = t.nextSection
		t.sections[0].onPrevSection = t.prevSection
		t.sections[0].isSelected = func() bool { return t.selection == 0 }

		t.sections[1].maxCol = 2
		t.sections[1].maxValue = 59
		t.sections[1].onNextSection = t.nextSection
		t.sections[1].onPrevSection = t.prevSection
		t.sections[1].isSelected = func() bool { return t.selection == 1 }

		idx := 2
		if t.showSeconds {
			t.sections[idx].maxCol = 2
			t.sections[idx].maxValue = 59
			t.sections[idx].onNextSection = t.nextSection
			t.sections[idx].onPrevSection = t.prevSection
			t.sections[idx].isSelected = func() bool { return t.selection == idx }
			idx++
		}

		if t.clock12Hour {
			t.sections[0].maxValue = 11
			t.sections[0].onChange = func(old, new int) {
				if old == 11 && new == 0 {
					// 11 -> 0
					t.sections[idx].value = (t.sections[idx].value + 1) % 2
				} else if old == 0 && new == 11 {
					t.sections[idx].value = (t.sections[idx].value + 1) % 2
				}
			}

			t.sections[idx].maxCol = 1
			t.sections[idx].maxValue = 1
			t.sections[idx].onNextSection = t.nextSection
			t.sections[idx].onPrevSection = t.prevSection
			t.sections[idx].isSelected = func() bool { return t.selection == idx }
		}
	}
}

func (t *TimeInput) nextSection() {
	t.section++
	if t.section >= len(t.sections) {
		t.section = 0
		t.selection = 0
		t.sections[t.section].col = t.sections[t.section].maxCol
		defer fyne.CurrentApp().Driver().CanvasForObject(t).FocusNext()
	} else {
		t.selection = t.section
		t.sections[t.section].col = t.sections[t.section].maxCol
	}
}

func (t *TimeInput) onWheelScroll(e *fyne.ScrollEvent) {
	if t.Disabled() {
		return
	}
	t.propertyLock.Lock()
	if t.showFocus {
		s := t.getSection(e.Position)
		if s >= 0 && s < t.numericFields {
			if e.Scrolled.DY > 0 {
				t.sections[s].decrement()
			} else {
				t.sections[s].increment()
			}
			defer t.Refresh()
		}
	}
	t.propertyLock.Unlock()
}

// pasteFromClipboard inserts text from the clipboard content,
// starting from the cursor position.
func (t *TimeInput) pasteFromClipboard(clipboard fyne.Clipboard) {
	tm, err := time.Parse("15:4:5", clipboard.Content())
	if err != nil {
		return
	}
	year, month, day := t.clock.Date()
	hour, min, sec := tm.Clock()
	t.Set(time.Date(year, month, day, hour, min, sec, 0, t.clock.Location()))
}

func (t *TimeInput) prevSection(selectValue bool) {
	if t.section == 0 {
		t.sections[t.section].col = 2
		return
	}
	t.section--
	t.sections[t.section].col = t.sections[t.section].maxCol
	if selectValue {
		t.selection = t.section
	} else {
		t.selection = -1
	}
}

func (t *TimeInput) registerShortcut() {
	t.shortcut.AddShortcut(&fyne.ShortcutCopy{}, func(se fyne.Shortcut) {
		cpy := se.(*fyne.ShortcutCopy)
		t.copyToClipboard(cpy.Clipboard)
	})
	t.shortcut.AddShortcut(&fyne.ShortcutPaste{}, func(se fyne.Shortcut) {
		paste := se.(*fyne.ShortcutPaste)
		t.pasteFromClipboard(paste.Clipboard)
	})
}

func (t *TimeInput) startDoubleTapTimer() {
	var ctx context.Context
	t.tapLock.Lock()
	// cancel any previous routine
	if t.tapCancelFunc != nil {
		t.tapCancelFunc()
	}
	ctx, t.tapCancelFunc = context.WithDeadline(context.TODO(), time.Now().Add(time.Millisecond*doubleClickDelay))
	defer t.tapCancelFunc()
	t.tapLock.Unlock()

	<-ctx.Done()

	t.tapLock.Lock()
	defer t.tapLock.Unlock()

	t.tapCancelFunc = nil
	t.tapLast = 0
}

func (t *TimeInput) unSelect(clear bool) {
	if t.selection >= 0 {
		if clear && t.selection < t.numericFields {
			t.sections[t.selection].value = 0
		}
		t.sections[t.selection].col = t.sections[t.selection].value / 10
		t.selection = -1
	}
}

func (t *TimeInput) updateSections() {
	t.sections[1].value = t.clock.Minute()
	idx := 2
	if t.showSeconds {
		t.sections[idx].value = t.clock.Second()
		idx++
	}
	h := t.clock.Hour()
	if t.clock12Hour {
		if h >= 12 {
			t.sections[0].value = h - 12
			t.sections[idx].value = 1
		} else {
			t.sections[0].value = h
			t.sections[idx].value = 0
		}
	} else {
		t.sections[0].value = h
	}
}

func (t *TimeInput) valid() bool {
	if !t.sections[0].valid() || !t.sections[1].valid() {
		return false
	}
	return !t.showSeconds || t.showSeconds && t.sections[2].valid()
}
