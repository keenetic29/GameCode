package api

import (
	"log"
	"net/http"
	"server/internal/domain"
	"server/internal/repository"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	repo    *repository.GameRepository
	xmlRepo *repository.XMLRepository
}

func NewHandler(repo *repository.GameRepository, xmlRepo *repository.XMLRepository) *Handler {
	return &Handler{
		repo:    repo,
		xmlRepo: xmlRepo,
	}
}

func (h *Handler) CreateGame(c *gin.Context) {
	maxPlayers := 2
	if mp := c.PostForm("max_players"); mp != "" {
		if n, err := strconv.Atoi(mp); err == nil && n >= 2 && n <= 4 {
			maxPlayers = n
		}
	}

	creatorName := c.PostForm("creator_name")
	if creatorName == "" {
		creatorName = "Creator"
	}

	game := h.repo.CreateGame(maxPlayers)
	creator := domain.NewPlayer(creatorName)
	game.AddPlayer(creator)

	c.JSON(http.StatusOK, gin.H{
		"game_id":   game.ID,
		"player_id": creator.ID,
	})
}

func (h *Handler) JoinGame(c *gin.Context) {
    gameID := c.PostForm("game_id")
    name := c.PostForm("name")

    game, exists := h.repo.GetGame(gameID)
    if !exists {
        c.JSON(http.StatusNotFound, gin.H{"error": "Game not found"})
        return
    }

    // Проверяем можно ли присоединиться
    if !game.IsJoinable() {
        c.JSON(http.StatusForbidden, gin.H{
            "error": "Game is already finished or full",
        })
        return
    }

    player := domain.NewPlayer(name)
    if !h.repo.AddPlayer(gameID, player) {
        c.JSON(http.StatusForbidden, gin.H{"error": "Failed to join game"})
        return
    }

    game.StartIfReady()

    c.JSON(http.StatusOK, gin.H{
        "player_id": player.ID,
        "started":   game.Started,
    })
}

func (h *Handler) MakeGuess(c *gin.Context) {
	gameID := c.PostForm("game_id")
	playerID := c.PostForm("player_id")
	guess := c.PostForm("guess")

	game, exists := h.repo.GetGame(gameID)
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "Game not found"})
		return
	}

	if len(guess) != 4 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Guess must be 4 digits"})
		return
	}

	black, white, isWinner, gameOver := game.CheckGuess(playerID, guess)

	if gameOver {
        game.IsFinished = true
        game.FinishedAt = time.Now()
        if isWinner {
            game.Winner = playerID
        }
        
        // Сохраняем перед отправкой ответа
        if err := h.xmlRepo.SaveGameResult(game); err != nil {
            log.Printf("Failed to save game result: %v", err)
        }
    }

	c.JSON(http.StatusOK, gin.H{
		"black":    black,
		"white":    white,
		"isWinner": isWinner,
		"gameOver": gameOver,
	})
}

func (h *Handler) GameStatus(c *gin.Context) {
    gameID := c.Param("id")
    game, exists := h.repo.GetGame(gameID)
    if !exists {
        c.JSON(http.StatusNotFound, gin.H{"error": "Game not found"})
        return
    }

    c.JSON(http.StatusOK, gin.H{
        "started":     game.Started,
        "finished":    game.IsFinished, // Добавляем статус завершения
        "players":     len(game.Players),
        "max_players": game.MaxPlayers,
    })
}