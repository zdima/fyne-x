package widget

import (
	"fmt"
	"strconv"
	"unicode"

	"fyne.io/fyne/v2"
)

type section struct {
	value         int
	col           int
	maxValue      int
	maxCol        int
	pos           fyne.Position
	size          fyne.Size
	enterZero     bool
	onNextSection func()
	onPrevSection func(selectValue bool)
	isSelected    func() bool
	onChange      func(old, new int)
}

func (s *section) decrement() {
	old := s.value
	s.value--
	if s.value < 0 {
		s.value = s.maxValue
	}
	if s.onChange != nil {
		s.onChange(old, s.value)
	}
}

func (s *section) increment() {
	old := s.value
	s.value++
	if s.value > s.maxValue {
		s.value = 0
	}
	if s.onChange != nil {
		s.onChange(old, s.value)
	}
}

func (s *section) get() string {
	if s.maxValue == 1 {
		switch s.value {
		case 0:
			return "AM"
		default:
			return "PM"
		}
	}
	if s.enterZero {
		return "0"
	}
	if s.maxValue == 11 && s.value == 0 {
		return strconv.Itoa(12)
	}
	if s.maxValue > 24 {
		return fmt.Sprintf("%02d", s.value)
	}
	return strconv.Itoa(s.value)
}

func (s *section) keyPressed(key *fyne.KeyEvent) bool {
	s.enterZero = false
	switch key.Name {
	case fyne.KeyLeft:
		s.col = s.maxCol
		s.onPrevSection(true)
		return true

	case fyne.KeyRight:
		s.col = s.maxCol
		s.onNextSection()
		return true

	case fyne.KeyUp:
		s.increment()
		return true

	case fyne.KeyDown:
		s.decrement()
		return true
	}
	return false
}

func (s *section) runePressed(r rune) bool {
	old := s.value
	if s.maxCol == 1 {
		// for AM/PM
		switch r {
		case 'a', 'A':
			s.value = 0
		case 'p', 'P':
			s.value = 1
		case ' ':
			s.value = (s.value + 1) % 2
		default:
			return false
		}
		if s.onChange != nil {
			s.onChange(old, s.value)
		}
		return true
	}
	if !unicode.IsDigit(r) {
		if r == '.' {
			s.col = s.maxCol
			s.onNextSection()
			return true
		}
		return false
	}
	s.value *= 10
	s.value += int(r - '0')
	if s.maxValue == 11 && s.value == 12 {
		s.value = 0
	}
	s.enterZero = s.value == 0 && r == '0'
	s.col++
	if s.onChange != nil {
		s.onChange(old, s.value)
	}
	if s.col == s.maxCol {
		s.onNextSection()
	}
	return true
}

func (s *section) valid() bool {
	return s.value >= 0 && s.value <= s.maxValue
}
