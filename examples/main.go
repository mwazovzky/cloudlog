package main

import (
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"time"

	"github.com/mwazovzky/cloudlog"
	"github.com/mwazovzky/cloudlog/config"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load configuration: %v\n", err)
		return
	}

	// Create a new HTTP client
	httpClient := &http.Client{}
	client := cloudlog.NewClient(cfg.LokiURL, cfg.LokiUsername, cfg.LokiAuthToken, httpClient)

	// Create a new logger
	logger := cloudlog.NewLogger(client, "cloudlog")

	rng := rand.New(rand.NewSource(time.Now().UnixNano()))

	for i := 0; i < 10; i++ {
		name := randomStr(8, rng)
		age := rng.Intn(100)

		// Log a message
		err = logger.Info("Welcome to Cloudlog!", "source", "Code", "Name", name, "Age", age)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to log message: %v\n", err)
		}
	}
}

func randomStr(n int, rng *rand.Rand) string {
	letters := []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
	result := make([]rune, n)
	for i := range result {
		result[i] = letters[rng.Intn(len(letters))]
	}
	return string(result)
}
