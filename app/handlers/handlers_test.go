package handlers_test

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
	"github.com/capgainschristian/go_api_ds/routes"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
)

func TestMain(m *testing.M) {

	database.ConnectDb()

	cache.ConnectRedis()

	code := m.Run()

	os.Exit(code)
}

func TestSignUp(t *testing.T) {
	if database.DB.Db == nil {
		t.Fatal(("Database is not initialized"))
	}

	result := database.DB.Db.Exec("DELETE FROM users WHERE email = ?", "admin@grahamsummitllc.com")
	if result.Error != nil {
		t.Fatal("Failed to delete from customers:", result.Error)
	}

	user := &models.User{
		Email:    "admin@grahamsummitllc.com",
		Password: "thisissecured",
	}

	jsonUser, _ := json.Marshal(user)

	router := routes.SetupRouter(cache.RedisClient.Client)

	req, err := http.NewRequest("POST", "/signup", bytes.NewBuffer(jsonUser))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusAccepted, rr.Code)
	expected := "User added successfully."
	assert.Equal(t, expected, rr.Body.String())

	var count int64
	database.DB.Db.Model(models.User{}).Where("email = ?", user.Email).Count(&count)
	assert.Equal(t, int64(1), count)
}

func TestLogin(t *testing.T) {
	if database.DB.Db == nil {
		t.Fatal("Database is not initialized")
	}

	result := database.DB.Db.Exec("SELECT FROM users WHERE email = ?", "admin@grahamsummitllc.com")
	if result.Error != nil {
		t.Fatal("User does not exist:", result.Error)
	}

	user := &models.User{
		Email:    "admin@grahamsummitllc.com",
		Password: "thisissecured",
	}

	jsonUser, _ := json.Marshal(user)

	router := routes.SetupRouter(cache.RedisClient.Client)

	req, err := http.NewRequest("POST", "/login", bytes.NewBuffer(jsonUser))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	expected := "Authentication was successful."
	assert.Equal(t, expected, rr.Body.String())

}

func TestAddCustomer(t *testing.T) {
	if database.DB.Db == nil {
		t.Fatal("Database is not initialized")
	}

	_, err := cache.RedisClient.Client.Ping(context.Background()).Result()
	if err != nil {
		log.Fatalf("Redis is not running: %v", err)
	}

	// Ensure the account doesn't exist already.
	result := database.DB.Db.Exec("DELETE FROM customers WHERE email = ?", "christian.graham@grahamsummitllc.com")
	if result.Error != nil {
		t.Fatal("Failed to delete from customers:", result.Error)
	}

	user := &models.User{
		Email:    "admin@grahamsummitllc.com",
		Password: "thisissecured",
	}

	customer := &models.Customer{
		Name:    "Christian Graham",
		Email:   "christian.graham@grahamsummitllc.com",
		Address: "777 Summit LLC Drive",
		Number:  1111,
	}

	jsonCustomer, _ := json.Marshal(customer)

	router := routes.SetupRouter(cache.RedisClient.Client)

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": user.Email,
		"exp": time.Now().Add(time.Hour * 24 * 30).Unix(), // 30 days expiration
	})

	tokenString, err := token.SignedString([]byte(os.Getenv("BCRYPT_KEY")))
	if err != nil {
		t.Fatal("Failed to sign token:", err)
	}

	req, err := http.NewRequest("POST", "/addcustomer", bytes.NewBuffer(jsonCustomer))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(&http.Cookie{
		Name:     "token",
		Value:    tokenString,
		Path:     "/",
		Expires:  time.Now().Add(time.Hour * 24 * 30),
		HttpOnly: true,
		Secure:   false,
	})

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

func TestUpdateCustomer(t *testing.T) {
	if database.DB.Db == nil {
		t.Fatal("Database is not initialized")
	}

	user := &models.User{
		Email:    "admin@grahamsummitllc.com",
		Password: "thisissecured",
	}

	// Update the customer
	updatedCustomer := &models.Customer{
		Name:    "Christian Updated",
		Email:   "christian.graham@grahamsummitllc.com",
		Address: "888 Summit LLC Drive",
		Number:  2222,
	}

	jsonCustomer, _ := json.Marshal(updatedCustomer)

	router := routes.SetupRouter(cache.RedisClient.Client)

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": user.Email,
		"exp": time.Now().Add(time.Hour * 24 * 30).Unix(), // 30 days expiration
	})

	tokenString, err := token.SignedString([]byte(os.Getenv("BCRYPT_KEY")))
	if err != nil {
		t.Fatal("Failed to sign token:", err)
	}

	req, err := http.NewRequest("PUT", "/updatecustomer", bytes.NewBuffer(jsonCustomer))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(&http.Cookie{
		Name:     "token",
		Value:    tokenString,
		Path:     "/",
		Expires:  time.Now().Add(time.Hour * 24 * 30),
		HttpOnly: true,
		Secure:   false,
	})

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	expected := "Customer's information updated successfully."
	assert.Equal(t, expected, rr.Body.String())

	// Verify customer was updated
	var customer models.Customer
	err = database.DB.Db.Where("email = ?", updatedCustomer.Email).First(&customer).Error
	assert.NoError(t, err)
	assert.Equal(t, updatedCustomer.Name, customer.Name)
	assert.Equal(t, updatedCustomer.Address, customer.Address)
	assert.Equal(t, updatedCustomer.Number, customer.Number)
}

func TestDeleteCustomer(t *testing.T) {

	user := &models.User{
		Email:    "admin@grahamsummitllc.com",
		Password: "thisissecured",
	}

	customer := &models.Customer{
		Name:    "Christian Graham",
		Email:   "christian.graham@grahamsummitllc.com",
		Address: "777 Summit LLC Drive",
		Number:  1111,
	}

	jsonCustomer, _ := json.Marshal(map[string]interface{}{
		"id":    customer.ID,
		"email": customer.Email,
	})

	router := routes.SetupRouter(cache.RedisClient.Client)

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": user.Email,
		"exp": time.Now().Add(time.Hour * 24 * 30).Unix(), // 30 days expiration
	})

	tokenString, err := token.SignedString([]byte(os.Getenv("BCRYPT_KEY")))
	if err != nil {
		t.Fatal("Failed to sign token:", err)
	}

	req, err := http.NewRequest("DELETE", "/deletecustomer", bytes.NewBuffer(jsonCustomer))
	if err != nil {
		t.Fatal()
	}
	req.Header.Set("Content-Type", "applicaton/json")
	req.AddCookie(&http.Cookie{
		Name:     "token",
		Value:    tokenString,
		Path:     "/",
		Expires:  time.Now().Add(time.Hour * 24 * 30),
		HttpOnly: true,
		Secure:   false,
	})

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
