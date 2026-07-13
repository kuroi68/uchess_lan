package uchess

import "testing"

func TestInputAppendAndLengthRespectsMaxLength(t *testing.T) {
	in := NewInput()

	for i := 0; i < MaxLength+5; i++ {
		in.Append('x')
	}

	if got := in.Length(); got != MaxLength {
		t.Fatalf("expected Length() == MaxLength (%d), got %d", MaxLength, got)
	}
}

func TestInputBackspaceWithMultibyteRune(t *testing.T) {
	in := NewInput()

	// Append a multi-byte rune (such as '😀') and a single-byte rune.
	in.Append('😀')
	in.Append('x')

	// Length is in bytes, not runes. '😀' is four bytes in UTF-8, plus one byte for 'x'.
	if got := in.Length(); got != 5 {
		t.Fatalf("expected byte length 5 after appends, got %d", got)
	}

	in.Backspace()
	if got := in.Length(); got != 4 {
		t.Fatalf("expected byte length 4 after first backspace (emoji only), got %d", got)
	}

	in.Backspace()
	if got := in.Length(); got != 0 {
		t.Fatalf("expected byte length 0 after second backspace, got %d", got)
	}
}

func TestInputCurrentPadsToMaxLength(t *testing.T) {
	in := NewInput()
	in.Append('a')
	in.Append('b')

	cur := in.Current()
	if len(cur) != MaxLength {
		t.Fatalf("expected Current() length %d, got %d", MaxLength, len(cur))
	}
	if cur[:2] != "ab" {
		t.Fatalf("expected Current() to start with \"ab\", got %q", cur)
	}
}

func TestInputClearResetsBuffer(t *testing.T) {
	in := NewInput()
	in.Append('z')

	in.Clear()
	if got := in.Length(); got != 0 {
		t.Fatalf("expected Length() == 0 after Clear(), got %d", got)
	}
}


