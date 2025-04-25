package domain

import "math/rand"

type Player struct {
	ID   string
	Name string
}

func NewPlayer(name string) Player {
	return Player{
		ID:   generatePlayerID(),
		Name: name,
	}
}

func generatePlayerID() string {
	const chars = "abcdefghijkmnopqrstuvwxyz0123456789"
	b := make([]byte, 8)
	for i := range b {
		b[i] = chars[rand.Intn(len(chars))]
	}
	return string(b)
}