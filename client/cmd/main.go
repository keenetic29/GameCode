package main

import (
	"client/internal/console"
)

func main() {
	client := console.NewConsoleClient()
	client.Run()
}