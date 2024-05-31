package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/capgainschristian/go_api_ds/database"
	"github.com/capgainschristian/go_api_ds/models"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
)

func TestMain(m *testing.M) {
	database.ConnectDb()

	code := m.Run()

	os.Exit(code)
}

func TestAddCustomer(t *testing.T) {
	// This wipes the database first before running the test - future improvement is to spawn a different db for testing
	database.DB.Db.Exec("DELETE FROM customers")

	router := mux.NewRouter()
	router.HandleFunc("/customer", AddCustomer).Methods("POST")

	customer := &models.Customer{
		Name:    "Christian Graham",
		Email:   "christian.graham@grahamsummitllc.com",
		Address: "777 Summit LLC Drive",
		Number:  1111,
	}

	jsonCustomer, _ := json.Marshal(customer)

	req, err := http.NewRequest("POST", "/customer", bytes.NewBuffer(jsonCustomer))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusAccepted, rr.Code)
	expected := "Customer added successfully."
	assert.Equal(t, expected, rr.Body.String())

	// Verify customer was added
	var count int64
	database.DB.Db.Model(models.Customer{}).Where("email = ?", customer.Email).Count(&count)
	assert.Equal(t, int64(1), count)
}

func TestDeleteCustomer(t *testing.T) {
	// This wipes the database first before running the test - future improvement is to spawn a different db for testing
	database.DB.Db.Exec("DELETE FROM customers")

	customer := &models.Customer{
		Name:    "Christian Graham",
		Email:   "christian.graham@grahamsummitllc.com",
		Address: "777 Summit LLC Drive",
		Number:  1111,
	}

	database.DB.Db.Create(&customer)

	router := mux.NewRouter()
	router.HandleFunc("/customer", AddCustomer).Methods("DELETE")

	jsonCustomer, _ := json.Marshal(map[string]interface{}{
		"id":    customer.ID,
		"email": customer.Email,
	})

	req, err := http.NewRequest("DELETE", "/customer", bytes.NewBuffer(jsonCustomer))
	if err != nil {
		t.Fatal()
	}
	req.Header.Set("Content-Type", "applicaton/json")

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	expected := "Customer deleted successfully."
	assert.Equal(t, expected, rr.Body.String())

	// Verify customer was deleted
	var count int64
	database.DB.Db.Model(&models.Customer{}).Where("email = ?", customer.Email).Count(&count)
	assert.Equal(t, int64(0), count)
}
