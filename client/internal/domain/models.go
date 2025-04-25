package domain

type GameState struct {
	GameID    string
	PlayerID  string
	Started   bool
	Black     int
	White     int
	IsWinner  bool
	GameOver  bool
}