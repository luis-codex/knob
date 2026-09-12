package preferences

// VolumeStep bounds. A step below the minimum makes the volume keys inert;
// above the maximum a single press swings across the range.
const (
	MinVolumeStep     = 1
	MaxVolumeStep     = 50
	DefaultVolumeStep = 5
)

// VolumeStep is how much a left/right press moves the volume, in percent.
type VolumeStep struct {
	value int
}

func NewVolumeStep(value int) (VolumeStep, error) {
	if value < MinVolumeStep || value > MaxVolumeStep {
		return VolumeStep{}, ErrInvalidVolumeStep
	}
	return VolumeStep{value: value}, nil
}

func (v VolumeStep) Value() int { return v.value }

// Audio groups the audio preferences.
type Audio struct {
	VolumeStep VolumeStep
}

func defaultAudio() Audio {
	return Audio{VolumeStep: VolumeStep{value: DefaultVolumeStep}}
}
