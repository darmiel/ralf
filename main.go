package main

import (
	"context"
	"github.com/go-redis/redis/v9"
	"github.com/ralf-life/engine/server"
)

func main() {
	// connect to redis
	var rc *redis.Client

	// use mock client for development
	rc = redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	if err := rc.Ping(context.TODO()).Err(); err != nil {
		panic(err)
	}

	demo := server.New(rc)
	if err := demo.Start(); err != nil {
		panic(err)
	}
}
