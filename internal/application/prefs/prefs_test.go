package prefs_test

import (
	"context"
	"errors"
	"testing"

	"knob/internal/application/prefs"
	"knob/internal/domain/preferences"
)

// fakeRepository is a Repository double: no disk, one recorded interaction
// per method, an injectable failure.
type fakeRepository struct {
	loaded  preferences.Loaded
	loadErr error

	saved   preferences.Settings
	savedOK bool
	saveErr error
}

func (f *fakeRepository) Load(ctx context.Context) (preferences.Loaded, error) {
	return f.loaded, f.loadErr
}

func (f *fakeRepository) Save(ctx context.Context, s preferences.Settings) error {
	if f.saveErr != nil {
		return f.saveErr
	}
	f.saved, f.savedOK = s, true
	return nil
}

func TestLoadReturnsWhatTheRepositoryHas(t *testing.T) {
	want := preferences.Default()
	want.Audio.VolumeStep, _ = preferences.NewVolumeStep(20)
	repo := &fakeRepository{loaded: preferences.Loaded{
		Settings: want,
		Rejected: []preferences.Rejected{{Key: "theme.mode", Reason: "boom"}},
	}}

	got, err := prefs.NewLoad(prefs.Deps{Repo: repo}).Execute(context.Background(), prefs.LoadCommand{})
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if got.Settings != want {
		t.Errorf("Settings = %+v, want %+v", got.Settings, want)
	}
	if len(got.Rejected) != 1 || got.Rejected[0].Key != "theme.mode" {
		t.Errorf("Rejected = %+v", got.Rejected)
	}
}

func TestLoadPropagatesTheRepositoryError(t *testing.T) {
	wantErr := errors.New("disk on fire")
	repo := &fakeRepository{loadErr: wantErr}

	if _, err := prefs.NewLoad(prefs.Deps{Repo: repo}).Execute(context.Background(), prefs.LoadCommand{}); !errors.Is(err, wantErr) {
		t.Errorf("error = %v, want %v", err, wantErr)
	}
}

func TestSavePersistsThroughTheRepository(t *testing.T) {
	repo := &fakeRepository{}
	want := preferences.Default()
	want.Interface.SidebarHidden = true

	got, err := prefs.NewSave(prefs.Deps{Repo: repo}).Execute(context.Background(), prefs.SaveCommand{Settings: want})
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if !repo.savedOK || repo.saved != want {
		t.Errorf("repository received %+v (ok=%v), want %+v", repo.saved, repo.savedOK, want)
	}
	if got.Settings != want {
		t.Errorf("response = %+v, want %+v", got.Settings, want)
	}
}

func TestSavePropagatesTheRepositoryError(t *testing.T) {
	wantErr := errors.New("syntax error in config.toml")
	repo := &fakeRepository{saveErr: wantErr}

	if _, err := prefs.NewSave(prefs.Deps{Repo: repo}).Execute(context.Background(), prefs.SaveCommand{Settings: preferences.Default()}); !errors.Is(err, wantErr) {
		t.Errorf("error = %v, want %v", err, wantErr)
	}
}
