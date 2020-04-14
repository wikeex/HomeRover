package mode

type TransMode bool

type Trans struct {
	Cmd   TransMode
	Video TransMode
	Audio TransMode
}