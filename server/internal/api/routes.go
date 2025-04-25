package api

import (
	"github.com/gin-gonic/gin"
)

type Router struct {
	handler *Handler
}

func NewRouter(handler *Handler) *gin.Engine {
	r := gin.Default()

	r.POST("/create", handler.CreateGame)
	r.POST("/join", handler.JoinGame)
	r.POST("/guess", handler.MakeGuess)
	r.GET("/game/:id/status", handler.GameStatus)

	return r
}