package main

import (
	"server/internal/api"
	"server/internal/repository"
)

func main() {
	gameRepo := repository.NewGameRepository()
	xmlRepo := repository.NewXMLRepository("game_results.xml")
	handler := api.NewHandler(gameRepo, xmlRepo)
	router := api.NewRouter(handler)
	
	router.Run(":8080")
}