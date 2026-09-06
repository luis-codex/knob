package pages

import (
	"testing"

	"knob/internal/domain/audio"
	"knob/internal/shared/components"
)

// audioFixture builds an Audio page with the given row counts per section and
// the list counts already synced, as HandleMsg would leave them after a load.
func audioFixture(outputs, inputs, streams int) *Audio {
	p := &Audio{
		lists: map[section]*components.List{
			sectionOutputs: components.NewList(),
			sectionInputs:  components.NewList(),
			sectionStreams: components.NewList(),
		},
		outputs: make([]audio.Device, outputs),
		inputs:  make([]audio.Device, inputs),
		streams: make([]audio.Stream, streams),
	}
	p.list(sectionOutputs).SetCount(outputs)
	p.list(sectionInputs).SetCount(inputs)
	p.list(sectionStreams).SetCount(streams)
	return p
}

func TestMoveCursorWithinSection(t *testing.T) {
	p := audioFixture(3, 2, 0)
	p.focused = sectionOutputs

	p.moveCursor(+1)
	if got := p.list(sectionOutputs).Cursor(); got != 1 {
		t.Fatalf("down: cursor = %d, want 1", got)
	}
	p.moveCursor(-1)
	p.moveCursor(-1) // already at the top: stays put, no crossing
	if got := p.list(sectionOutputs).Cursor(); got != 0 {
		t.Fatalf("up past top: cursor = %d, want 0", got)
	}
	if p.focused != sectionOutputs {
		t.Fatalf("up past top: focused = %d, want sectionOutputs", p.focused)
	}
}

func TestMoveCursorCrossesSections(t *testing.T) {
	p := audioFixture(2, 2, 0)
	p.focused = sectionOutputs
	p.list(sectionOutputs).ToLast()

	p.moveCursor(+1) // off the bottom of Outputs -> top of Microphones
	if p.focused != sectionInputs {
		t.Fatalf("cross down: focused = %d, want sectionInputs", p.focused)
	}
	if got := p.list(sectionInputs).Cursor(); got != 0 {
		t.Fatalf("cross down: cursor = %d, want 0", got)
	}

	p.moveCursor(-1) // back off the top -> bottom of Outputs
	if p.focused != sectionOutputs {
		t.Fatalf("cross up: focused = %d, want sectionOutputs", p.focused)
	}
	if got := p.list(sectionOutputs).Cursor(); got != 1 {
		t.Fatalf("cross up: cursor = %d, want 1 (last row)", got)
	}
}

func TestMoveCursorSkipsEmptySection(t *testing.T) {
	p := audioFixture(2, 0, 2) // Microphones empty
	p.focused = sectionOutputs
	p.list(sectionOutputs).ToLast()

	p.moveCursor(+1)
	if p.focused != sectionStreams {
		t.Fatalf("focused = %d, want sectionStreams (skipping empty Microphones)", p.focused)
	}
	if got := p.list(sectionStreams).Cursor(); got != 0 {
		t.Fatalf("cursor = %d, want 0", got)
	}
}

func TestMoveCursorBottomEdgeInert(t *testing.T) {
	p := audioFixture(2, 0, 0)
	p.focused = sectionOutputs
	p.list(sectionOutputs).ToLast()

	p.moveCursor(+1)
	if p.focused != sectionOutputs {
		t.Fatalf("focused = %d, want sectionOutputs", p.focused)
	}
	if got := p.list(sectionOutputs).Cursor(); got != 1 {
		t.Fatalf("cursor = %d, want 1", got)
	}
}

func TestJumpSectionSkipsEmptyAndWraps(t *testing.T) {
	p := audioFixture(2, 0, 2) // Microphones empty
	p.focused = sectionOutputs

	p.jumpSection(+1)
	if p.focused != sectionStreams {
		t.Fatalf("tab: focused = %d, want sectionStreams", p.focused)
	}

	p.focused = sectionOutputs
	p.jumpSection(-1)
	if p.focused != sectionStreams {
		t.Fatalf("shift+tab: focused = %d, want sectionStreams (wrap, skip empty)", p.focused)
	}
}

func TestJumpSectionKeepsCursorAndStaysWhenAlone(t *testing.T) {
	p := audioFixture(2, 0, 2)
	p.list(sectionStreams).ToLast() // cursor at 1

	p.focused = sectionOutputs
	p.jumpSection(+1)
	if got := p.list(sectionStreams).Cursor(); got != 1 {
		t.Fatalf("cursor = %d, want 1 (jump keeps each section's cursor)", got)
	}

	p = audioFixture(2, 0, 0)
	p.focused = sectionOutputs
	p.jumpSection(+1)
	if p.focused != sectionOutputs {
		t.Fatalf("focused = %d, want sectionOutputs (nowhere else to go)", p.focused)
	}
}
