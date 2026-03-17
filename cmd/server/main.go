package main

import (
	"fmt"
	"log"

	"viperai/internal/config"
	"viperai/internal/domain"
	"viperai/internal/infrastructure/cache"
	"viperai/internal/infrastructure/database"
	"viperai/internal/infrastructure/queue"
	"viperai/internal/repository"
	"viperai/internal/transport/http/router"
)

func main() {
	_ = config.Load()

	if err := initializeInfrastructure(); err != nil {
		log.Fatalf("Failed to initialize infrastructure: %v", err)
	}

	cfg := config.Get()
	addr := fmt.Sprintf("%s:%d", cfg.App.Host, cfg.App.Port)
	log.Printf("Starting %s server on %s", cfg.App.Name, addr)

	r := router.Setup()
	if err := r.Run(addr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

func initializeInfrastructure() error {
	if err := database.Initialize(); err != nil {
		return fmt.Errorf("database initialization failed: %w", err)
	}
	log.Println("Database initialized successfully")

	if err := cache.Initialize(); err != nil {
		return fmt.Errorf("cache initialization failed: %w", err)
	}
	log.Println("Cache initialized successfully")

	queue.StartConsumer(saveMessageToDatabase)
	log.Println("Message queue initialized successfully")

	return nil
}

func saveMessageToDatabase(msg *domain.ChatMessage) error {
	msgRepo := repository.NewMessageRepository(database.GetDB())
	return msgRepo.Create(msg)
}
