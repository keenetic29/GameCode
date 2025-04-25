package repository

import (
	"server/internal/domain"
	"sync"
)

type GameRepository struct {
	games map[string]*domain.Game
	mu    sync.RWMutex
}

func NewGameRepository() *GameRepository {
	return &GameRepository{
		games: make(map[string]*domain.Game),
	}
}

func (r *GameRepository) CreateGame(maxPlayers int) *domain.Game {
	r.mu.Lock()
	defer r.mu.Unlock()

	game := domain.NewGame(maxPlayers)
	r.games[game.ID] = game
	return game
}

func (r *GameRepository) GetGame(gameID string) (*domain.Game, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	game, exists := r.games[gameID]
	return game, exists
}

func (r *GameRepository) AddPlayer(gameID string, player domain.Player) bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	game, exists := r.games[gameID]
	if !exists {
		return false
	}

	return game.AddPlayer(player)
}