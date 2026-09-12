package pages

import (
	"context"
	"errors"
	"testing"

	"knob/internal/application/prefs"
	"knob/internal/domain/preferences"
)

// fakeRepository is a Repository double: no disk, one recorded Save, an
// injectable failure.
type fakeRepository struct {
	saved   preferences.Settings
	savedOK bool
	saveErr error
}

func (f *fakeRepository) Load(ctx context.Context) (preferences.Loaded, error) {
	return preferences.Loaded{Settings: preferences.Default()}, nil
}

func (f *fakeRepository) Save(ctx context.Context, s preferences.Settings) error {
	if f.saveErr != nil {
		return f.saveErr
	}
	f.saved, f.savedOK = s, true
	return nil
}

// preferencesFixture builds a Preferences page with the cursor already on r
// and the list's count synced, as a View call would leave it. rows and the
// row constants share the same order, so the constant doubles as an index.
func preferencesFixture(repo *fakeRepository, r row) *Preferences {
	uc := prefs.NewUseCases(prefs.Deps{Repo: repo})
	p := NewPreferences(context.Background(), "Settings", uc, preferences.Default())
	p.list.SetCount(len(rows))
	for i := 0; i < int(r); i++ {
		p.list.Next()
	}
	return p
}

func TestAdjustVolumeStepClampsAtBounds(t *testing.T) {
	repo := &fakeRepository{}
	p := preferencesFixture(repo, rowVolumeStep)

	p.settings.Audio.VolumeStep, _ = preferences.NewVolumeStep(preferences.MinVolumeStep)
	if cmd := p.adjust(-1); cmd != nil {
		t.Fatalf("adjust(-1) at the minimum: got a save command, want nil")
	}
	if got := p.settings.Audio.VolumeStep.Value(); got != preferences.MinVolumeStep {
		t.Errorf("volume step = %d, want unchanged at %d", got, preferences.MinVolumeStep)
	}

	p.settings.Audio.VolumeStep, _ = preferences.NewVolumeStep(preferences.MaxVolumeStep)
	if cmd := p.adjust(+1); cmd != nil {
		t.Fatalf("adjust(+1) at the maximum: got a save command, want nil")
	}
	if got := p.settings.Audio.VolumeStep.Value(); got != preferences.MaxVolumeStep {
		t.Errorf("volume step = %d, want unchanged at %d", got, preferences.MaxVolumeStep)
	}
}

func TestAdjustThemeModeCyclesWithoutWrapping(t *testing.T) {
	repo := &fakeRepository{}
	p := preferencesFixture(repo, rowThemeMode)

	if got := p.settings.Theme.Mode.String(); got != preferences.ModeAuto {
		t.Fatalf("default mode = %s, want %s", got, preferences.ModeAuto)
	}

	p.adjust(-1) // already at the first entry: clamp, no wrap to dark
	if got := p.settings.Theme.Mode.String(); got != preferences.ModeAuto {
		t.Errorf("mode = %s, want still %s (clamped at the first entry)", got, preferences.ModeAuto)
	}

	p.adjust(+1) // auto -> light
	p.adjust(+1) // light -> dark
	if got := p.settings.Theme.Mode.String(); got != preferences.ModeDark {
		t.Fatalf("mode = %s, want %s", got, preferences.ModeDark)
	}

	p.adjust(+1) // already at the last entry: clamp, no wrap to auto
	if got := p.settings.Theme.Mode.String(); got != preferences.ModeDark {
		t.Errorf("mode = %s, want still %s (clamped at the last entry)", got, preferences.ModeDark)
	}
}

func TestAdjustBooleansToggleRegardlessOfDirection(t *testing.T) {
	repo := &fakeRepository{}
	p := preferencesFixture(repo, rowAnimations)
	want := !p.settings.Interface.Animations

	p.adjust(-1)
	if p.settings.Interface.Animations != want {
		t.Fatalf("animations = %v, want %v after a left press", p.settings.Interface.Animations, want)
	}
	p.adjust(-1)
	if p.settings.Interface.Animations == want {
		t.Fatalf("animations = %v, want it flipped back", p.settings.Interface.Animations)
	}

	p2 := preferencesFixture(repo, rowSidebarHidden)
	wantSidebar := !p2.settings.Interface.SidebarHidden

	p2.adjust(+1)
	if p2.settings.Interface.SidebarHidden != wantSidebar {
		t.Fatalf("sidebar_hidden = %v, want %v after a right press", p2.settings.Interface.SidebarHidden, wantSidebar)
	}
}

// TestAdjustAppliesBeforeTheSaveRoundTripCompletes guards against a
// regression where adjust read p.settings without having written its own
// previous change to it: two presses before the async PreferencesSavedMsg
// from the first one arrives would both compute from the same starting
// value and the second press would be silently dropped.
func TestAdjustAppliesBeforeTheSaveRoundTripCompletes(t *testing.T) {
	repo := &fakeRepository{}
	p := preferencesFixture(repo, rowVolumeStep)
	start := p.settings.Audio.VolumeStep.Value()

	p.adjust(+1)
	p.adjust(+1)

	if got := p.settings.Audio.VolumeStep.Value(); got != start+2 {
		t.Fatalf("volume step = %d, want %d (two presses, no confirmation in between)", got, start+2)
	}
}

func TestAdjustSavesThroughTheRepository(t *testing.T) {
	repo := &fakeRepository{}
	p := preferencesFixture(repo, rowVolumeStep)

	cmd := p.adjust(+1)
	if cmd == nil {
		t.Fatalf("adjust: want a save command")
	}
	msg := cmd()
	saved, ok := msg.(PreferencesSavedMsg)
	if !ok {
		t.Fatalf("cmd() = %T, want PreferencesSavedMsg", msg)
	}
	if !repo.savedOK || repo.saved != saved.Settings {
		t.Errorf("repository received %+v (ok=%v), want %+v", repo.saved, repo.savedOK, saved.Settings)
	}
}

func TestAdjustReportsASaveFailure(t *testing.T) {
	repo := &fakeRepository{saveErr: errors.New("syntax error in config.toml")}
	p := preferencesFixture(repo, rowVolumeStep)

	cmd := p.adjust(+1)
	if cmd == nil {
		t.Fatalf("adjust: want a save command even when the repository will fail")
	}
	if _, ok := cmd().(preferencesFailedMsg); !ok {
		t.Fatalf("cmd() = %T, want preferencesFailedMsg", cmd())
	}
}
