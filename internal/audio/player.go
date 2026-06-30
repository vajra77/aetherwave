package audio

type Player struct {
}

func New() (*Player, error) {
	return &Player{}, nil
}

func (p *Player) Play(samples Sample) error {
	return nil
}
