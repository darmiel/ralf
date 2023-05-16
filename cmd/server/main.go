package main

import (
	"context"
	"fmt"
	"github.com/redis/go-redis/v9"
	"github.com/ralf-life/engine/internal/server"
)

var (
	version string
	commit  string
	date    string
)

func main() {
	fmt.Println("starting engine", version, "commit:", commit, "at", date)
	// connect to redis
	var rc *redis.Client

	// use mock client for development
	rc = redis.NewClient(&redis.Options{
		Addr:     "redis:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	if err := rc.Ping(context.TODO()).Err(); err != nil {
		panic(err)
	}

	demo := server.New(rc, version, commit, date)
	if err := demo.Start(); err != nil {
		panic(err)
	}
}
