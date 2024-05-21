package routes

import (
	"github.com/capgainschristian/go_api_ds/handlers"
	"github.com/gorilla/mux"
)

func SetupRouter() *mux.Router {
	r := mux.NewRouter()

	r.HandleFunc("/healthcheck", handlers.HealthCheck).Methods("GET")
	// Technically, all of the routes below can be "/customers" since there are different VERBS for different functions
	r.HandleFunc("/listcustomers", handlers.ListCustomers).Methods("GET")
	r.HandleFunc("/addcustomer", handlers.AddCustomer).Methods("POST")
	r.HandleFunc("/deletecustomer", handlers.DeleteCustomer).Methods("DELETE")
	r.HandleFunc("/updatecustomer", handlers.UpdateCustomer).Methods("PUT")

	return r
}
