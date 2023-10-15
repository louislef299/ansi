package main

import (
	"errors"

	"github.com/Azure/go-ansiterm"
)

type stringHandler struct {
	ansiterm.AnsiEventHandler
	cursor int
	b      []byte
}

func (h *stringHandler) Print(b byte) error {
	if h.cursor == len(h.b) {
		h.b = append(h.b, b)
	} else {
		h.b[h.cursor] = b
	}
	h.cursor++
	return nil
}

func (h *stringHandler) Execute(b byte) error {
	switch b {
	case '\b':
		if h.cursor > 0 {
			if h.cursor == len(h.b) && h.b[h.cursor-1] == ' ' {
				h.b = h.b[:len(h.b)-1]
			}
			h.cursor--
		}
	case '\r', '\n':
		h.Print(b)
	}
	return nil
}

// Erase Display
func (h *stringHandler) ED(v int) error {
	switch v {
	case 1: // Erase from start to cursor.
		for i := 0; i < h.cursor; i++ {
			h.b[i] = ' '
		}
	case 2, 3: // Erase whole display.
		h.b = make([]byte, h.cursor)
		for i := range h.b {
			h.b[i] = ' '
		}
	default: // Erase from cursor to end of display.
		h.b = h.b[:h.cursor+1]
	}
	return nil
}

// CUrsor Position
func (h *stringHandler) CUP(x, y int) error {
	if x > 1 {
		return errors.New("termtest: cursor position not supported for X > 1")
	}
	if y > len(h.b) {
		for n := len(h.b) - y; n > 0; n-- {
			h.b = append(h.b, ' ')
		}
	}
	h.cursor = y - 1
	return nil
}

func (h *stringHandler) String() string {
	return string(h.b)
}
