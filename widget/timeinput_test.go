package widget

import (
	"testing"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/test"
	"github.com/stretchr/testify/assert"
)

func TestTimeInput_SetGet(t *testing.T) {
	entry := NewTimeInput(false)
	v := time.Now()
	entry.Set(v)
	assert.Equal(t, v.Round(time.Minute), entry.Get(), "Get not returning same value after Set using minutes")
}

func TestTimeInput_SetGetSeconds(t *testing.T) {
	entry := NewTimeInput(true)
	v := time.Now()
	entry.Set(v)
	assert.Equal(t, v.Round(time.Second), entry.Get(), "Get not returning same value after Set using seconds")
}

func TestTimeInput_Entry24(t *testing.T) {

	Clock12Hour = false

	entry := NewTimeInput(true)
	win := test.NewWindow(entry)
	win.Resize(fyne.NewSize(500, 300))
	defer win.Close()

	tm := entry.Get()

	win.Canvas().Focus(entry)
	win.Canvas().Focused().TypedRune('1')
	win.Canvas().Focused().TypedRune('5')
	win.Canvas().Focused().TypedRune('3')
	win.Canvas().Focused().TypedRune('2')
	win.Canvas().Focused().TypedRune('4')
	win.Canvas().Focused().TypedRune('5')

	year, month, day := tm.Date()
	e := time.Date(year, month, day, 15, 32, 45, 0, tm.Location())

	assert.Equal(t, e, entry.Get())
}

func TestTimeInput_Entry12(t *testing.T) {

	Clock12Hour = true

	entry := NewTimeInput(true)
	win := test.NewWindow(entry)
	win.Resize(fyne.NewSize(500, 300))
	defer win.Close()

	tm := entry.Get()

	win.Canvas().Focus(entry)
	win.Canvas().Focused().TypedRune('1')
	win.Canvas().Focused().TypedRune('0')
	win.Canvas().Focused().TypedRune('3')
	win.Canvas().Focused().TypedRune('2')
	win.Canvas().Focused().TypedRune('4')
	win.Canvas().Focused().TypedRune('5')
	win.Canvas().Focused().TypedRune('p')

	year, month, day := tm.Date()
	e := time.Date(year, month, day, 22, 32, 45, 0, tm.Location())

	assert.Equal(t, e, entry.Get())
}

func TestTimeInput_Focus12(t *testing.T) {

	Clock12Hour = true

	entry := NewTimeInput(true)
	win := test.NewWindow(entry)
	defer win.Close()

	tm := entry.Get()
	canvas := win.Canvas()
	pos := entry.Position()
	size := entry.Size()
	segmentWidth := size.Width / 4

	tap := pos.Add(fyne.NewDelta(segmentWidth/2, size.Height/2))
	test.TapCanvas(canvas, tap)
	assert.Equal(t, entry, win.Canvas().Focused())

	win.Canvas().Focused().TypedKey(&fyne.KeyEvent{Name: fyne.KeyUp})
	assert.Equal(t, tm.Hour()+1, entry.Get().Hour())

	tap = pos.Add(fyne.NewDelta(segmentWidth+segmentWidth/2, size.Height/2))
	test.TapCanvas(canvas, tap)
	assert.Equal(t, entry, win.Canvas().Focused())

	win.Canvas().Focused().TypedKey(&fyne.KeyEvent{Name: fyne.KeyUp})
	m := tm.Minute() + 1
	if m >= 60 {
		m = 0
	}
	assert.Equal(t, m, entry.Get().Minute(), "minutes")

	tap = pos.Add(fyne.NewDelta(segmentWidth*2+segmentWidth/2, size.Height/2))
	test.TapCanvas(canvas, tap)
	assert.Equal(t, entry, win.Canvas().Focused())

	win.Canvas().Focused().TypedKey(&fyne.KeyEvent{Name: fyne.KeyUp})
	m = tm.Second() + 1
	if m >= 60 {
		m = 0
	}
	assert.Equal(t, m, entry.Get().Second(), "seconds")

	tap = pos.Add(fyne.NewDelta(segmentWidth*3+segmentWidth/2, size.Height/2))
	test.TapCanvas(canvas, tap)
	assert.Equal(t, entry, win.Canvas().Focused())

	win.Canvas().Focused().TypedKey(&fyne.KeyEvent{Name: fyne.KeyUp})
	m = tm.Hour() + 1 + 12 // 1 hour added above, 12 hours added by KeyUp
	if m >= 24 {
		m -= 24
	}
	assert.Equal(t, m, entry.Get().Hour())

	// flip back with second (double) tap
	time.Sleep(time.Millisecond * 100)
	test.TapCanvas(canvas, tap)
	m = tm.Hour() + 1 // 1 hour added above
	if m >= 24 {
		m -= 24
	}
	assert.Equal(t, m, entry.Get().Hour())
}

func TestTimeInput_Range12(t *testing.T) {

	Clock12Hour = true

	entry := NewTimeInput(true)
	win := test.NewWindow(entry)
	defer win.Close()

	year, month, day := time.Now().Date()

	tm := time.Date(year, month, day, 10, 58, 58, 0, time.Now().Location())
	entry.Set(tm)

	canvas := win.Canvas()
	pos := entry.Position()
	size := entry.Size()
	segmentWidth := size.Width / 4

	// move to hours
	tap := pos.Add(fyne.NewDelta(segmentWidth/2, size.Height/2))
	test.TapCanvas(canvas, tap)
	assert.Equal(t, entry, win.Canvas().Focused())
	// up
	for i := 0; i < 15; i++ {
		h := (48 + (11 + i)) % 24
		win.Canvas().Focused().TypedKey(&fyne.KeyEvent{Name: fyne.KeyUp})
		assert.Equal(t, h, entry.Get().Hour(), "hour +%d", i)
	}
	// down
	for i := 0; i < 26; i++ {
		h := (48 - i) % 24
		win.Canvas().Focused().TypedKey(&fyne.KeyEvent{Name: fyne.KeyDown})
		assert.Equal(t, h, entry.Get().Hour(), "hour -%d", i)
	}

	// move to minutes
	tap = pos.Add(fyne.NewDelta(segmentWidth+segmentWidth/2, size.Height/2))
	test.TapCanvas(canvas, tap)
	assert.Equal(t, entry, win.Canvas().Focused())
	// up
	for i := 0; i < 80; i++ {
		m := (120 + 59 + i) % 60
		win.Canvas().Focused().TypedKey(&fyne.KeyEvent{Name: fyne.KeyUp})
		assert.Equal(t, m, entry.Get().Minute(), "minute +%d", i)
	}
	// down
	for i := 0; i < 80; i++ {
		m := (120 + 19 + i) % 60
		win.Canvas().Focused().TypedKey(&fyne.KeyEvent{Name: fyne.KeyUp})
		assert.Equal(t, m, entry.Get().Minute(), "minute -%d", i)
	}

	// move to seconds
	tap = pos.Add(fyne.NewDelta(segmentWidth*2+segmentWidth/2, size.Height/2))
	test.TapCanvas(canvas, tap)
	assert.Equal(t, entry, win.Canvas().Focused())
	// up
	for i := 0; i < 80; i++ {
		m := (120 + 59 + i) % 60
		win.Canvas().Focused().TypedKey(&fyne.KeyEvent{Name: fyne.KeyUp})
		assert.Equal(t, m, entry.Get().Second(), "second +%d", i)
	}
	// down
	for i := 0; i < 80; i++ {
		m := (120 + 19 + i) % 60
		win.Canvas().Focused().TypedKey(&fyne.KeyEvent{Name: fyne.KeyUp})
		assert.Equal(t, m, entry.Get().Second(), "second -%d", i)
	}
}

func TestTimeInput_Range24(t *testing.T) {

	Clock12Hour = false

	entry := NewTimeInput(true)
	win := test.NewWindow(entry)
	defer win.Close()

	year, month, day := time.Now().Date()

	tm := time.Date(year, month, day, 10, 58, 58, 0, time.Now().Location())
	entry.Set(tm)

	canvas := win.Canvas()
	pos := entry.Position()
	size := entry.Size()
	segmentWidth := size.Width / 4

	// move to hours
	tap := pos.Add(fyne.NewDelta(segmentWidth/2, size.Height/2))
	test.TapCanvas(canvas, tap)
	assert.Equal(t, entry, win.Canvas().Focused())
	// up
	for i := 0; i < 15; i++ {
		h := (48 + (11 + i)) % 24
		win.Canvas().Focused().TypedKey(&fyne.KeyEvent{Name: fyne.KeyUp})
		assert.Equal(t, h, entry.Get().Hour(), "hour +%d", i)
	}
	// down
	for i := 0; i < 26; i++ {
		h := (48 - i) % 24
		win.Canvas().Focused().TypedKey(&fyne.KeyEvent{Name: fyne.KeyDown})
		assert.Equal(t, h, entry.Get().Hour(), "hour -%d", i)
	}
}
