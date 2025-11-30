package main

import (
	"log"

	"gin-mongo-aws/internal/config"
	"gin-mongo-aws/internal/route"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	s := server.NewServer(cfg)
	s.Run()
}
