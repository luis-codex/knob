package pulse

import (
	"bufio"
	"context"
	"os/exec"
	"strings"
	"syscall"
	"time"

	"settings-cli/internal/domain/audio"
	"settings-cli/internal/domain/errs"
)

// watched is the set of objects whose events trigger a re-read.
//
// "client" is left out: every pactl invocation -- every re-read included --
// opens and closes a client, so listening to it would feed the loop back into
// itself. sink-input events do matter: they are the mixer streams, and they
// are born, change or die when an application starts, pauses or stops playing.
// Listing them with pactl creates none, so there is no feedback.
var watched = map[string]bool{
	"sink":       true,
	"source":     true,
	"sink-input": true, // mixer streams
	"server":     true, // changes when the default changes
	"card":       true, // plugging or unplugging a device
}

// Watcher listens to `pactl subscribe`.
type Watcher struct{}

var _ audio.Watcher = (*Watcher)(nil)

func NewWatcher() *Watcher { return &Watcher{} }

func (w *Watcher) Changes(ctx context.Context) (<-chan struct{}, error) {
	cmd := exec.CommandContext(ctx, binary, "subscribe")

	// `pactl subscribe` does not end on its own: without cancelling the
	// context it is orphaned when the application closes.
	//
	// Pdeathsig is not used because it is not reliable in Go: it fires when
	// the thread that launched the process dies, and the runtime moves
	// goroutines between threads. WaitDelay ensures the SIGKILL if the
	// SIGTERM is not enough.
	cmd.Cancel = func() error { return cmd.Process.Signal(syscall.SIGTERM) }
	cmd.WaitDelay = 2 * time.Second

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, errs.Wrap(errs.KindConflict, "could not listen to the sound server", err)
	}
	if err := cmd.Start(); err != nil {
		return nil, errs.Wrap(errs.KindConflict, "could not listen to the sound server", err)
	}

	// Capacity 1: notifications coalesce rather than queue. A burst of events
	// must cause one re-read, not twenty.
	changes := make(chan struct{}, 1)

	go func() {
		defer close(changes)
		defer func() { _ = cmd.Wait() }()

		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			if !relevant(scanner.Text()) {
				continue
			}

			select {
			case changes <- struct{}{}:
			case <-ctx.Done():
				return
			default: // a notification is already pending: no need to send another
			}
		}
	}()

	return changes, nil
}

// relevant decides whether a line "Event 'change' on sink #58" matters.
func relevant(line string) bool {
	_, target, ok := strings.Cut(line, " on ")
	if !ok {
		return false
	}

	// "sink #58" -> "sink". The token is kept whole: "sink" and "sink-input"
	// are different objects and are checked separately in watched.
	object, _, _ := strings.Cut(strings.TrimSpace(target), " ")
	return watched[object]
}
