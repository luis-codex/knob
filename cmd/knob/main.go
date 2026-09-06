// SPDX-License-Identifier: GPL-3.0-or-later

package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io/fs"
	"os"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"

	"knob/internal/app"
	"knob/internal/application/devices"
	"knob/internal/application/sound"
	"knob/internal/config"
	"knob/internal/domain/bluetooth"
	"knob/internal/infrastructure/bluez"
	"knob/internal/infrastructure/memory"
	"knob/internal/infrastructure/pulse"
	simulatedbt "knob/internal/infrastructure/simulated"
)

// scanWindow is how long a device discovery lasts.
const scanWindow = 8 * time.Second

// Exit codes. A script wrapping knob can branch on them; they are
// documented in the README.
//
//	0   success
//	1   runtime failure
//	2   bad arguments (set by the flag package)
//	3   stdout is not an interactive terminal
//	130 interrupted with Ctrl-C
const (
	exitError      = 1
	exitNoTerminal = 3
)

// Build metadata. In a release the linker injects it
// (-ldflags "-X main.version=… -X main.commit=… -X main.date=…"); a plain
// build keeps these values.
var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	flag.Usage = usage
	simulated := flag.Bool("fake-bluetooth", false, "use fake Bluetooth devices instead of the system ones (for development)")
	writeTheme := flag.Bool("write-theme", false, "write an example theme to the config directory and exit")
	showVersion := flag.Bool("version", false, "print the version and exit")
	flag.Parse()

	if *showVersion {
		fmt.Printf("knob %s (commit %s, %s)\n", version, commit, date)
		return
	}

	if *writeTheme {
		writeExampleTheme()
		return
	}

	// Composition root: the only place that knows the concrete
	// implementations. Choosing a backend means changing these lines, nothing
	// else. The context is cancelled on exit: that is what releases the
	// processes listening to the system.
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// A broken theme must not stop startup: it is reported and we carry on
	// with the default colors.
	theme, themeErrs := config.LoadTheme()
	for _, err := range themeErrs {
		fmt.Fprintln(os.Stderr, "theme:", err)
	}

	deviceSvc := devices.NewService(bluetoothPorts(*simulated))
	soundSvc := sound.NewService(pulse.NewRepository(), pulse.NewStreamRepository(), pulse.NewWatcher())

	if _, err := tea.NewProgram(app.New(ctx, theme, deviceSvc, soundSvc)).Run(); err != nil {
		// The most common startup failure is having no interactive terminal
		// (a pipe, CI, cron). Explain it instead of dumping the raw library
		// error.
		if isNoTTY(err) {
			fmt.Fprintln(os.Stderr, "knob needs an interactive terminal; it does not work over a pipe or without a TTY.")
			os.Exit(exitNoTerminal)
		}
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(exitError)
	}
}

// isNoTTY recognizes bubbletea's failure when there is no terminal. It matches
// on text because the library wraps the OS error without exporting a sentinel
// for this case.
func isNoTTY(err error) bool {
	return strings.Contains(err.Error(), "TTY")
}

// usage explains what knob is and how to run it. The flag package appends the
// list of flags below via PrintDefaults.
func usage() {
	_, _ = fmt.Fprint(flag.CommandLine.Output(), `knob — system settings (audio, Bluetooth, network) in a TUI.

Usage:
  knob [flags]

With no flags it opens the interface. It needs an interactive terminal.

Examples:
  knob                     open the settings
  knob -write-theme        drop an example theme in ~/.config/knob/
  knob -fake-bluetooth     use fake devices (development)
  knob -version            print the version

Issues and questions:
  https://github.com/luis-codex/knob/issues

Flags:
`)
	flag.PrintDefaults()
}

// bluetoothPorts returns the three Bluetooth ports.
//
// By default they are the system ones: a settings tool must show what is
// really there. The fake ones sit behind a flag, for developing without
// hardware.
func bluetoothPorts(fake bool) (bluetooth.Repository, bluetooth.AdapterRepository, bluetooth.Scanner) {
	if !fake {
		return bluez.NewRepository(), bluez.NewAdapterRepository(), bluez.NewScanner(scanWindow)
	}

	repo := memory.NewDeviceRepository()
	seedDevices(context.Background(), repo)
	return repo, memory.NewAdapterRepository(true), simulatedbt.NewScanner(scanWindow)
}

// seedDevices loads example devices into the simulated backend. They are built
// through the service so they go through the same validations.
func seedDevices(ctx context.Context, repo bluetooth.Repository) {
	svc := devices.NewService(repo, memory.NewAdapterRepository(true), simulatedbt.NewScanner(0))

	seed := []struct {
		address, name, kind string
		pair, connect       bool
		battery             int
	}{
		{"AA:BB:CC:DD:EE:FF", "WH-1000XM4", "headphones", true, true, 82},
		{"11:22:33:44:55:66", "MX Master 3S", "mouse", true, true, 45},
		{"77:88:99:AA:BB:CC", "Keyboard K380", "keyboard", true, false, 0},
	}

	for _, d := range seed {
		report(svc.Discover(ctx, d.address, d.name, d.kind))
		if d.pair {
			report(svc.Pair(ctx, d.address))
		}
		if d.connect {
			report(svc.Connect(ctx, d.address))
			report(svc.ReportBattery(ctx, d.address, d.battery))
		}
	}
}

// report drops the value and reports the error: a seeding failure must not
// stop startup, but it must not go unnoticed either.
func report[T any](_ T, err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, "seed:", err)
	}
}

// writeExampleTheme leaves the theme template in place and reports where.
func writeExampleTheme() {
	path, err := config.WriteExampleTheme()
	switch {
	case errors.Is(err, fs.ErrExist):
		fmt.Fprintf(os.Stderr, "a theme already exists at %s; leaving it untouched\n", path)
		os.Exit(1)
	case err != nil:
		fmt.Fprintln(os.Stderr, "could not write the theme:", err)
		os.Exit(1)
	}
	fmt.Println("example theme written to", path)
}
