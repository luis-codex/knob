package pulse

import (
	"context"
	"encoding/json"
	"strconv"

	"settings-cli/internal/domain/audio"
	"settings-cli/internal/domain/errs"
)

// StreamRepository expone los flujos de audio de las aplicaciones.
type StreamRepository struct{}

var _ audio.StreamRepository = (*StreamRepository)(nil)

func NewStreamRepository() *StreamRepository { return &StreamRepository{} }

// jsonStream es un sink-input de pactl.
type jsonStream struct {
	Index      int                    `json:"index"`
	Mute       bool                   `json:"mute"`
	Corked     bool                   `json:"corked"`
	Volume     map[string]jsonChannel `json:"volume"`
	Properties map[string]any         `json:"properties"`
}

// property lee una propiedad de texto. El mapa viene con tipos mezclados y
// las ausentes llegan como null.
func (s jsonStream) property(key string) string {
	if value, ok := s.Properties[key].(string); ok {
		return value
	}
	return ""
}

func (r *StreamRepository) List(ctx context.Context) ([]audio.Stream, error) {
	raw, err := run(ctx, "list", "sink-inputs")
	if err != nil {
		return nil, err
	}

	var decoded []jsonStream
	if err := json.Unmarshal(raw, &decoded); err != nil {
		return nil, errs.Wrap(errs.KindConflict, "respuesta ilegible del servidor de sonido", err)
	}

	streams := make([]audio.Stream, 0, len(decoded))
	for _, s := range decoded {
		if stream, ok := toStream(s); ok {
			streams = append(streams, stream)
		}
	}
	return streams, nil
}

func (r *StreamRepository) FindByID(ctx context.Context, id audio.StreamID) (audio.Stream, error) {
	streams, err := r.List(ctx)
	if err != nil {
		return audio.Stream{}, err
	}

	for _, s := range streams {
		if s.ID().Equals(id) {
			return s, nil
		}
	}
	return audio.Stream{}, audio.ErrStreamNotFound
}

// Save aplica volumen y silencio. Ambos comandos son idempotentes.
func (r *StreamRepository) Save(ctx context.Context, s audio.Stream) error {
	if s.IsZero() {
		return audio.ErrInvalidStreamID
	}

	index := strconv.Itoa(s.ID().Index())

	if _, err := run(ctx, "set-sink-input-volume", index, strconv.Itoa(s.Volume().Level())+"%"); err != nil {
		return err
	}

	muted := "0"
	if s.Muted() {
		muted = "1"
	}
	_, err := run(ctx, "set-sink-input-mute", index, muted)
	return err
}

// toStream traduce un sink-input al dominio. Devuelve false si no describe un
// flujo utilizable.
func toStream(s jsonStream) (audio.Stream, bool) {
	id, err := audio.NewStreamID(s.Index)
	if err != nil {
		return audio.Stream{}, false
	}

	// application.name es lo que reconoce el usuario. Sin él se recurre al
	// binario y, en último caso, al índice: un flujo sin nombre debe salir
	// igualmente en el mezclador.
	app, err := audio.NewName(firstNonEmpty(
		s.property("application.name"),
		s.property("application.process.binary"),
		"flujo "+strconv.Itoa(s.Index),
	))
	if err != nil {
		return audio.Stream{}, false
	}

	volume, ok := volumeOf(s.Volume)
	if !ok {
		return audio.Stream{}, false
	}

	stream, err := audio.RestoreStream(id, app, s.property("media.name"), volume, s.Mute, s.Corked)
	if err != nil {
		return audio.Stream{}, false
	}
	return stream, true
}
