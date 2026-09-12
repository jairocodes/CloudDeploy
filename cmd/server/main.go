package main

import (
	"log"
	"os"

	"github.com/jairocodes/CloudDeploy.git/internal/api"
	"github.com/jairocodes/CloudDeploy.git/internal/api/handlers"
	"github.com/jairocodes/CloudDeploy.git/internal/queue"
	"github.com/jairocodes/CloudDeploy.git/internal/store"
	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("no .env file found, using existing environment varibles")
	}

	db, err := store.Connect(os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatalf("connect db: %v", err)
	}
	defer db.Close()

	rdb := redis.NewClient(&redis.Options{Addr: os.Getenv("REDIS_ADDR")})

	s := store.New(db)
	q := queue.New(rdb)
	h := handlers.New(s, q)

	r := api.NewRouter(h)
	if err := r.Run(":8080"); err != nil {
		log.Fatalf("server: %v", err)
	}
}

//ejemplos de prueba con curl
//"id":"b72712d1-f8ee-4b4f-b868-4fa3f9908b66" de prueba
//"deploy_id":"9e70c2d7-9be4-4b74-aabe-5764fa315a5d","sha":"a1b2c3d"
//deploys":null,"project_id":"b72712d1-f8ee-4b4f-b868-4fa3f9908b66
