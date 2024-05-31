package handlers

import (
	"testing"

)

func TestMain(m *testing.M) {
	database.ConnectDb

	code := m.Run()

	os.Exit(code)
}

func TestAddCustomer(t *testing.T) {
	router := mux.NewRouter()
	

}