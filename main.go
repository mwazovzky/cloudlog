package main

import (
	"log/slog"
	"math/rand"
	"time"
)

type Entry struct {
	Name string
	Age  int
}

func main() {
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))

	count := 1000

	for i := 0; i < count; i++ {
		name := randomString(8, rng)
		age := rng.Intn(100) // Random age between 0 and 99
		slog.Info("Generated Entry", "Name", name, "Age", age)
	}
}

func randomString(n int, rng *rand.Rand) string {
	letters := []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
	result := make([]rune, n)
	for i := range result {
		result[i] = letters[rng.Intn(len(letters))]
	}
	return string(result)
}
