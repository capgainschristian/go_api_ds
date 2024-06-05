// Need to add:
// Error handling for all.
// Testing. - Need to fix testing.
// Pagination for "all" retrieval.
// Redis/caching. - Need to add to update. Then introduce async to make it even faster.
// ENV vars/const.
// Authorization?

// CRUD: Create (Post), Read (Get), Update (Put), Delete (Delete)

package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/capgainschristian/go_api_ds/database"
	"github.com/capgainschristian/go_api_ds/routes"
	"github.com/go-redis/redis/v8"
)

const PORT = 3000

func main() {

	database.ConnectDb()

	rdb := redis.NewClient(&redis.Options{
		Addr:     "cache:6379",
		Password: "capgainschristian",
		DB:       0,
	})

	// Check Redis connectivity
	_, err := rdb.Ping(context.Background()).Result()
	if err != nil {
		log.Fatalf("Failed to connec to Redis: %v", err)
	}

	router := routes.SetupRouter(rdb)

	log.Printf("Server listening on :%d...\n", PORT)

	err = http.ListenAndServe(fmt.Sprintf(":%d", PORT), router)

	if err != nil {
		log.Fatal(err)
	}

}
