package widget

import (
	"image/color"
	"sync"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/theme"
)

const cursorInterruptTime = 300 * time.Millisecond

type EntryCursorAnimation struct {
	Cursor            *canvas.Rectangle
	mu                *sync.RWMutex
	anim              *fyne.Animation
	lastInterruptTime time.Time

	timeNow func() time.Time // useful for testing
}

func NewEntryCursorAnimation(cursor *canvas.Rectangle) *EntryCursorAnimation {
	a := &EntryCursorAnimation{mu: &sync.RWMutex{}, Cursor: cursor, timeNow: time.Now}
	return a
}

// creates fyne animation
func (a *EntryCursorAnimation) createAnim(inverted bool) *fyne.Animation {
	var cursorOpaque color.Color = theme.PrimaryColor()
	r, g, b, _ := theme.PrimaryColor().RGBA()
	var cursorDim color.Color = color.NRGBA{R: uint8(r), G: uint8(g), B: uint8(b), A: 0x16}
	start, end := cursorDim, cursorOpaque
	if inverted {
		start, end = cursorOpaque, cursorDim
	}
	interrupted := false
	anim := canvas.NewColorRGBAAnimation(start, end, time.Second/2, func(c color.Color) {
		a.mu.RLock()
		shouldInterrupt := a.timeNow().Sub(a.lastInterruptTime) <= cursorInterruptTime
		a.mu.RUnlock()
		if shouldInterrupt {
			if !interrupted {
				a.Cursor.FillColor = cursorOpaque
				a.Cursor.Refresh()
				interrupted = true
			}
			return
		}
		if interrupted {
			a.mu.Lock()
			a.anim.Stop()
			if !inverted {
				a.anim = a.createAnim(true)
			}
			interrupted = false
			a.mu.Unlock()
			go func() {
				a.mu.RLock()
				canStart := a.anim != nil
				a.mu.RUnlock()
				if canStart {
					a.anim.Start()
				}
			}()
			return
		}
		a.Cursor.FillColor = c
		a.Cursor.Refresh()
	})

	anim.RepeatCount = fyne.AnimationRepeatForever
	anim.AutoReverse = true
	return anim
}

// starts cursor animation.
func (a *EntryCursorAnimation) Start() {
	a.mu.Lock()
	isStopped := a.anim == nil
	if isStopped {
		a.anim = a.createAnim(false)
	}
	a.mu.Unlock()
	if isStopped {
		a.anim.Start()
	}
}

// temporarily stops the animation by "cursorInterruptTime".
func (a *EntryCursorAnimation) Interrupt() {
	a.mu.Lock()
	a.lastInterruptTime = a.timeNow()
	a.mu.Unlock()
}

// stops cursor animation.
func (a *EntryCursorAnimation) Stop() {
	a.mu.Lock()
	if a.anim != nil {
		a.anim.Stop()
		a.anim = nil
	}
	a.mu.Unlock()
}
