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
