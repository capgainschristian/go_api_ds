package routes

import (
	"net/http"

	"github.com/capgainschristian/go_api_ds/handlers"
	"github.com/go-redis/redis/v8"
	"github.com/gorilla/mux"
)

func SetupRouter(rdb *redis.Client) *mux.Router {
	r := mux.NewRouter()

	r.HandleFunc("/healthcheck", handlers.HealthCheck).Methods("GET")
	// Technically, all of the routes below can be "/customers" since there are different VERBS for different functions
	r.HandleFunc("/listcustomers", func(w http.ResponseWriter, r *http.Request) {
		handlers.ListCustomers(w, r, rdb)
	}).Methods("GET")
	r.HandleFunc("/addcustomer", func(w http.ResponseWriter, r *http.Request) {
		handlers.AddCustomer(w, r, rdb)
	}).Methods("POST")
	r.HandleFunc("/deletecustomer", handlers.DeleteCustomer).Methods("DELETE")
	r.HandleFunc("/updatecustomer", handlers.UpdateCustomer).Methods("PUT")

	return r
}
