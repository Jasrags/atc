package engine

import (
	"strings"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

// TextInput handles keyboard text input for the ATC command prompt.
type TextInput struct {
	text    string
	cursor  int
	focused bool
}

// NewTextInput creates a focused text input.
func NewTextInput() TextInput {
	return TextInput{focused: true}
}

// Update processes keyboard events and returns the submitted text (if Enter was pressed).
// Returns empty string if no submission.
func (t *TextInput) Update() string {
	if !t.focused {
		return ""
	}

	// Character input.
	chars := ebiten.AppendInputChars(nil)
	if len(chars) > 0 {
		before := t.text[:t.cursor]
		after := t.text[t.cursor:]
		t.text = before + string(chars) + after
		t.cursor += len(chars)
	}

	// Backspace.
	if inpututil.IsKeyJustPressed(ebiten.KeyBackspace) {
		if t.cursor > 0 {
			before := t.text[:t.cursor-1]
			after := t.text[t.cursor:]
			t.text = before + after
			t.cursor--
		}
	}

	// Delete.
	if inpututil.IsKeyJustPressed(ebiten.KeyDelete) {
		if t.cursor < len(t.text) {
			before := t.text[:t.cursor]
			after := t.text[t.cursor+1:]
			t.text = before + after
		}
	}

	// Home / End.
	if inpututil.IsKeyJustPressed(ebiten.KeyHome) && ebiten.IsKeyPressed(ebiten.KeyShift) {
		// Shift+Home is handled by camera; plain Home moves cursor.
	}
	if ebiten.IsKeyPressed(ebiten.KeyControl) || ebiten.IsKeyPressed(ebiten.KeyMeta) {
		if inpututil.IsKeyJustPressed(ebiten.KeyA) {
			t.cursor = 0
		}
		if inpututil.IsKeyJustPressed(ebiten.KeyE) {
			t.cursor = len(t.text)
		}
	}

	// Arrow keys for cursor movement (only when text is non-empty).
	if inpututil.IsKeyJustPressed(ebiten.KeyLeft) && t.cursor > 0 {
		t.cursor--
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyRight) && t.cursor < len(t.text) {
		t.cursor++
	}

	// Enter submits.
	if inpututil.IsKeyJustPressed(ebiten.KeyEnter) || inpututil.IsKeyJustPressed(ebiten.KeyNumpadEnter) {
		submitted := strings.TrimSpace(t.text)
		t.text = ""
		t.cursor = 0
		return submitted
	}

	return ""
}

// Text returns the current input text.
func (t *TextInput) Text() string {
	return t.text
}

// Cursor returns the cursor position.
func (t *TextInput) Cursor() int {
	return t.cursor
}

// IsEmpty reports whether the input field has no text.
// Uses raw length (not trimmed) so that a leading space does not trigger
// single-key shortcuts like p (freeze) or [ ] (speed).
func (t *TextInput) IsEmpty() bool {
	return len(t.text) == 0
}

// SetText sets the input text and moves cursor to the end.
func (t *TextInput) SetText(s string) {
	t.text = s
	t.cursor = len(s)
}
