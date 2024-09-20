package routes

import (
	"net/http"

	"github.com/capgainschristian/go_api_ds/handlers"
	"github.com/capgainschristian/go_api_ds/middleware"
	"github.com/gorilla/mux"
)

func SetupRouter() *mux.Router {
	r := mux.NewRouter()

	r.HandleFunc("/healthcheck", handlers.HealthCheck).Methods("GET")
	r.HandleFunc("/signup", handlers.SignUp).Methods("POST")
	r.HandleFunc("/login", handlers.Login).Methods("POST")
	r.HandleFunc("/customercreation", handlers.AddCustomer).Methods("POST")
	r.HandleFunc("/listcustomers", handlers.ListCustomers).Methods("GET")
	r.Handle("/addcustomer", middleware.AuthMiddleware(http.HandlerFunc(handlers.AddCustomer))).Methods("POST")
	r.Handle("/deletecustomer", middleware.AuthMiddleware(http.HandlerFunc(handlers.DeleteCustomer))).Methods("DELETE")
	r.Handle("/updatecustomer", middleware.AuthMiddleware(http.HandlerFunc(handlers.UpdateCustomer))).Methods("PUT")

	return r
}
