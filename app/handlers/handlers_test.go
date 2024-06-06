package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/capgainschristian/go_api_ds/cache"
	"github.com/capgainschristian/go_api_ds/database"
	"github.com/capgainschristian/go_api_ds/models"
	"github.com/go-redis/redis/v8"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
)

func TestMain(m *testing.M) {

	database.ConnectDb()

	cache.ConnectRedis()

	code := m.Run()

	os.Exit(code)
}

func setupRouter(rdb *redis.Client) *mux.Router {
	router := mux.NewRouter()
	router.HandleFunc("/addcustomer", func(w http.ResponseWriter, r *http.Request) {
		AddCustomer(w, r, rdb)
	}).Methods("POST")
	return router
}

func TestAddCustomer(t *testing.T) {
	if database.DB.Db == nil {
		t.Fatal("Database is not initialized")
	}

	_, err := cache.RedisClient.Client.Ping(context.Background()).Result()
	if err != nil {
		log.Fatalf("Redis is not running: %v", err)
	}

	// This wipes the database first before running the test - future improvement is to spawn a different db for testing
	result := database.DB.Db.Exec("DELETE FROM customers")
	if result.Error != nil {
		t.Fatal("Failed to delete from customers:", result.Error)
	}
	router := setupRouter(cache.RedisClient.Client)

	customer := &models.Customer{
		Name:    "Christian Graham",
		Email:   "christian.graham@grahamsummitllc.com",
		Address: "777 Summit LLC Drive",
		Number:  1111,
	}

	jsonCustomer, _ := json.Marshal(customer)

	req, err := http.NewRequest("POST", "/addcustomer", bytes.NewBuffer(jsonCustomer))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusAccepted, rr.Code)
	expected := "Customer added successfully."
	assert.Equal(t, expected, rr.Body.String())

	// Verify customer was added to the database
	var count int64
	database.DB.Db.Model(models.Customer{}).Where("email = ?", customer.Email).Count(&count)
	assert.Equal(t, int64(1), count)

	ctx := context.Background()
	cacheKey := "customer:" + customer.Email
	customerJSON, err := json.Marshal(customer)
	if err != nil {
		t.Fatalf("Failed to marshal customer: %v", err)
	}

	err = cache.RedisClient.Client.Set(ctx, cacheKey, customerJSON, 24*time.Hour).Err()
	if err != nil {
		t.Fatalf("Failed to add customer to the cache: %v", err)
	}

	// Verify customer was added to the cache
	cachedCustomer, err := cache.RedisClient.Client.Get(ctx, cacheKey).Result()
	assert.NoError(t, err)
	assert.NotEmpty(t, cachedCustomer)
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
	router.HandleFunc("/customer", DeleteCustomer).Methods("DELETE")

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
