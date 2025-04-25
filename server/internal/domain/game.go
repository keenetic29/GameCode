package domain

import (
	"encoding/xml"
	"math/rand"
	"time"
)

type Game struct {
	ID           string
	SecretCode   string
	Players      []Player
	MaxPlayers   int
	Current      int
	Attempts     map[string]int
	StartedAt    time.Time
	FinishedAt   time.Time
	Winner       string
	Started      bool
	IsFinished   bool
	MaxAttempts  int
}

type GameResult struct {
	XMLName     xml.Name  `xml:"game_result"`
	GameID      string    `xml:"game_id"`
	SecretCode  string    `xml:"secret_code"`
	StartedAt   time.Time `xml:"started_at"`
	FinishedAt  time.Time `xml:"finished_at"`
	Winner      string    `xml:"winner,omitempty"`
	Players     []PlayerResult `xml:"players>player"`
}

type PlayerResult struct {
	ID       string `xml:"id"`
	Name     string `xml:"name"`
	Attempts int    `xml:"attempts"`
}

func NewGame(maxPlayers int) *Game {
	return &Game{
		ID:          generateID(),
		SecretCode:  generateCode(),
		MaxPlayers:  maxPlayers,
		Players:     make([]Player, 0),
		Attempts:    make(map[string]int),
		StartedAt:   time.Now(),
		MaxAttempts: 5,
	}
}

func generateID() string {
	const chars = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"
	b := make([]byte, 6)
	for i := range b {
		b[i] = chars[rand.Intn(len(chars))]
	}
	return string(b)
}

func generateCode() string {
	b := make([]byte, 4)
	for i := range b {
		b[i] = byte(rand.Intn(10)) + '0'
	}
	return string(b)
}

func (g *Game) AddPlayer(player Player) bool {
	if len(g.Players) >= g.MaxPlayers {
		return false
	}
	g.Players = append(g.Players, player)
	g.Attempts[player.ID] = 0
	return true
}

func (g *Game) CheckGuess(playerID, guess string) (black, white int, isWinner, gameOver bool) {
	secret := g.SecretCode
	black, white = 0, 0
	secretMatched := make([]bool, 4)
	guessMatched := make([]bool, 4)

	// Check black markers
	for i := 0; i < 4; i++ {
		if secret[i] == guess[i] {
			black++
			secretMatched[i] = true
			guessMatched[i] = true
		}
	}

	// Check white markers
	for i := 0; i < 4; i++ {
		if guessMatched[i] {
			continue
		}
		for j := 0; j < 4; j++ {
			if !secretMatched[j] && guess[i] == secret[j] {
				white++
				secretMatched[j] = true
				break
			}
		}
	}

	g.Attempts[playerID]++
	g.Current = (g.Current + 1) % len(g.Players)

	isWinner = black == 4
	gameOver = isWinner || g.Attempts[playerID] >= g.MaxAttempts

	if gameOver {
        g.IsFinished = true
        g.FinishedAt = time.Now()
        if isWinner {
            g.Winner = playerID
        }
    }

    return black, white, isWinner, gameOver
}

func (g *Game) StartIfReady() bool {
	if len(g.Players) == g.MaxPlayers && !g.Started {
		g.Started = true
		return true
	}
	return false
}

func (g *Game) ToGameResult() GameResult {
    result := GameResult{
        GameID:     g.ID,
        SecretCode: g.SecretCode,
        StartedAt:  g.StartedAt,
        FinishedAt: g.FinishedAt,
        Winner:     g.Winner,
        Players:    make([]PlayerResult, 0, len(g.Players)), // Инициализируем слайс
    }

    // Добавляем только если есть игроки
    if g.Players != nil {
        for _, player := range g.Players {
            result.Players = append(result.Players, PlayerResult{
                ID:       player.ID,
                Name:     player.Name,
                Attempts: g.Attempts[player.ID],
            })
        }
    }

    return result
}

func (g *Game) IsJoinable() bool {
    return !g.IsFinished && len(g.Players) < g.MaxPlayers
}