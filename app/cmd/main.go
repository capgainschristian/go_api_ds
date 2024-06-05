// Need to add:
// Error handling for all.
// Testing.
// Pagination for "all" retrieval.
// Redis/caching.
// ENV vars/const.
// Async/go channel.
// Authorization?

// CRUD: Create (Post), Read (Get), Update (Put), Delete (Delete)

package main

import (
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
		Addr:     "localhost:6379",
		Password: "capgainschristian",
		DB:       0,
	})

	router := routes.SetupRouter(rdb)

	log.Printf("Server listening on :%d...\n", PORT)

	err := http.ListenAndServe(fmt.Sprintf(":%d", PORT), router)

	if err != nil {
		log.Fatal(err)
	}

}
