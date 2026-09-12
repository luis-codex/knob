// SPDX-License-Identifier: GPL-3.0-or-later

package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strings"

	tea "charm.land/bubbletea/v2"

	"knob/internal/app"
	"knob/internal/application/prefs"
	"knob/internal/application/sound"
	"knob/internal/infrastructure/pulse"
	"knob/internal/infrastructure/tomlstore"
)

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
	showVersion := flag.Bool("version", false, "print the version and exit")
	flag.Parse()

	if *showVersion {
		fmt.Printf("knob %s (commit %s, %s)\n", version, commit, date)
		return
	}

	// Composition root: the only place that knows the concrete
	// implementations. Choosing a backend means changing these lines, nothing
	// else. The context is cancelled on exit: that is what releases the
	// processes listening to the system.
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	prefsUC := prefs.NewUseCases(prefs.Deps{Repo: tomlstore.NewRepository()})

	// A store that does not parse, or a value it rejects, must not stop
	// startup: both are reported and knob carries on with the defaults.
	loaded, err := prefsUC.Load.Execute(ctx, prefs.LoadCommand{})
	if err != nil {
		fmt.Fprintln(os.Stderr, "config:", err)
	}
	for _, r := range loaded.Rejected {
		fmt.Fprintf(os.Stderr, "config: %s: %s\n", r.Key, r.Reason)
	}

	soundUC := sound.NewUseCases(sound.Deps{
		Repo:    pulse.NewRepository(),
		Streams: pulse.NewStreamRepository(),
		Watcher: pulse.NewWatcher(),
	})

	if _, err := tea.NewProgram(app.New(ctx, prefsUC, loaded.Settings, soundUC)).Run(); err != nil {
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
	_, _ = fmt.Fprint(flag.CommandLine.Output(), `knob — system settings (audio, network) in a TUI.

Usage:
  knob [flags]

With no flags it opens the interface. It needs an interactive terminal.

Examples:
  knob                     open the settings
  knob -version            print the version

Issues and questions:
  https://github.com/luis-codex/knob/issues

Flags:
`)
	flag.PrintDefaults()
}
