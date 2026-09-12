package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"

	"github.com/jairocodes/CloudDeploy.git/internal/queue"
	"github.com/jairocodes/CloudDeploy.git/internal/store"
	"github.com/jairocodes/CloudDeploy.git/internal/worker"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("no .env file found, using existing environment variables")
	}

	db, err := store.Connect(os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatalf("connect db: %v", err)
	}
	defer db.Close()

	rdb := redis.NewClient(&redis.Options{Addr: os.Getenv("REDIS_URL")})

	q := queue.New(rdb)
	s := store.New(db)
	w := worker.New(q, s)

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	w.Run(ctx)
	log.Println("worker shut down")
}
