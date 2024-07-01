package routes

import (
	"net/http"

	"github.com/capgainschristian/go_api_ds/handlers"
	"github.com/capgainschristian/go_api_ds/middleware"
	"github.com/go-redis/redis/v8"
	"github.com/gorilla/mux"
)

func SetupRouter(rdb *redis.Client) *mux.Router {
	r := mux.NewRouter()

	r.HandleFunc("/healthcheck", handlers.HealthCheck).Methods("GET")
	r.HandleFunc("/signup", handlers.SignUp).Methods("POST")
	r.HandleFunc("/login", handlers.Login).Methods("POST")
	r.HandleFunc("/customercreation", func(w http.ResponseWriter, r *http.Request) {
		handlers.AddCustomer(w, r, rdb)
	}).Methods("POST")
	r.HandleFunc("/listcustomers", func(w http.ResponseWriter, r *http.Request) {
		handlers.ListCustomers(w, r, rdb)
	}).Methods("GET")
	r.Handle("/addcustomer", middleware.AuthMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlers.AddCustomer(w, r, rdb)
	}))).Methods("POST")
	r.Handle("/deletecustomer", middleware.AuthMiddleware(http.HandlerFunc(handlers.DeleteCustomer))).Methods("DELETE")
	r.Handle("/updatecustomer", middleware.AuthMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlers.UpdateCustomer(w, r, rdb)
	}))).Methods("PUT")

	return r
}
