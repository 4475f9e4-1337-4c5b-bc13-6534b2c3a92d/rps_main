package game

import "github.com/google/uuid"

const (
	StatePlaying = iota
	StateRoundOver
	StateGameOver
)

const RoundTime = 5

type Score struct {
	PlayerOne int
	PlayerTwo int
	Draw      int
}

type Game struct {
	Id              string
	State           int
	Round           int
	Timer           int
	PlayerOneChoice string
	PlayerTwoChoice string
	Score           Score
}

func New() *Game {
	return &Game{
		Id:    uuid.NewString(),
		State: StatePlaying,
		Round: 1,
		Timer: RoundTime,
		Score: Score{},
	}
}
