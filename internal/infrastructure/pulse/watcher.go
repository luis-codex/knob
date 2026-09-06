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

// watched son los objetos cuyos eventos disparan una relectura.
//
// Queda fuera "client": cada invocación de pactl —cada relectura incluida—
// abre y cierra un cliente, así que escucharlo realimentaría el bucle. Los
// sink-input sí interesan: son los flujos del mezclador, y nacen, cambian o
// mueren cuando una aplicación empieza, pausa o para de sonar. Listarlos con
// pactl no crea ninguno, de modo que no hay realimentación.
var watched = map[string]bool{
	"sink":       true,
	"source":     true,
	"sink-input": true, // flujos del mezclador
	"server":     true, // cambia al cambiar el predeterminado
	"card":       true, // enchufar o quitar un dispositivo
}

// Watcher escucha `pactl subscribe`.
type Watcher struct{}

var _ audio.Watcher = (*Watcher)(nil)

func NewWatcher() *Watcher { return &Watcher{} }

func (w *Watcher) Changes(ctx context.Context) (<-chan struct{}, error) {
	cmd := exec.CommandContext(ctx, binary, "subscribe")

	// `pactl subscribe` no termina solo: sin cancelar el contexto queda
	// huérfano al cerrar la aplicación.
	//
	// No se usa Pdeathsig porque no es fiable en Go: se dispara al morir el
	// hilo que lanzó el proceso, y el runtime mueve las goroutines entre
	// hilos. WaitDelay asegura el SIGKILL si el SIGTERM no basta.
	cmd.Cancel = func() error { return cmd.Process.Signal(syscall.SIGTERM) }
	cmd.WaitDelay = 2 * time.Second

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, errs.Wrap(errs.KindConflict, "no se pudo escuchar al servidor de sonido", err)
	}
	if err := cmd.Start(); err != nil {
		return nil, errs.Wrap(errs.KindConflict, "no se pudo escuchar al servidor de sonido", err)
	}

	// Capacidad 1: los avisos se funden en vez de encolarse. Una ráfaga de
	// eventos debe provocar una relectura, no veinte.
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
			default: // ya hay un aviso pendiente: sobra mandar otro
			}
		}
	}()

	return changes, nil
}

// relevant decide si una línea "Event 'change' on sink #58" interesa.
func relevant(line string) bool {
	_, target, ok := strings.Cut(line, " on ")
	if !ok {
		return false
	}

	// "sink #58" -> "sink". El token va entero: "sink" y "sink-input" son
	// objetos distintos y se miran por separado en watched.
	object, _, _ := strings.Cut(strings.TrimSpace(target), " ")
	return watched[object]
}
