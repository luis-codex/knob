package preferences_test

import (
	"testing"

	"knob/internal/domain/preferences"
)

func TestDefaultInterface(t *testing.T) {
	got := preferences.Default().Interface
	if got.SidebarHidden {
		t.Error("default must start with the sidebar visible")
	}
	if !got.Animations {
		t.Error("default must start with animations on")
	}
}
