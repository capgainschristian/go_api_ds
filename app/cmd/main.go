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
	"fmt"
	"log"
	"net/http"

	"github.com/capgainschristian/go_api_ds/cache"
	"github.com/capgainschristian/go_api_ds/database"
	"github.com/capgainschristian/go_api_ds/routes"
)

const PORT = 3000

func main() {

	database.ConnectDb()

	cache.ConnectRedis()

	router := routes.SetupRouter()

	log.Printf("Server listening on :%d...\n", PORT)

	err := http.ListenAndServe(fmt.Sprintf(":%d", PORT), router)

	if err != nil {
		log.Fatal(err)
	}

}
