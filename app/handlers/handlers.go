package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/capgainschristian/go_api_ds/cache"
	"github.com/capgainschristian/go_api_ds/database"
	"github.com/capgainschristian/go_api_ds/models"
	"github.com/go-redis/redis/v8"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func HealthCheck(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)

	fmt.Fprint(w, "API is up and running.")
}

func SignUp(w http.ResponseWriter, r *http.Request) {
	newUser := new(models.User)

	err := json.NewDecoder(r.Body).Decode(&newUser)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if newUser.Email == "" {
		http.Error(w, "Missing user email", http.StatusBadRequest)
		return
	}

	if newUser.Password == "" {
		http.Error(w, "Missing user password", http.StatusBadRequest)
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(newUser.Password), 10)
	if err != nil {
		http.Error(w, "Failed to hash the password", http.StatusInternalServerError)
		return
	}

	// Need to convert byte to string to pass to DB
	newUser.Password = string(hash)

	err = database.DB.Db.Create(&newUser).Error
	if err != nil {
		http.Error(w, "Failed to add user to the database", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusAccepted)
	w.Write([]byte("User added successfully."))
}

func Login(w http.ResponseWriter, r *http.Request) {
	authReq := new(models.User)

	err := json.NewDecoder(r.Body).Decode(&authReq)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	user := new(models.User)

	if authReq.Email == "" {
		http.Error(w, "Need user email to query the database", http.StatusBadRequest)
		return
	} else {
		err = database.DB.Db.Unscoped().Where("email = ?", authReq.Email).First(&user).Error
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				http.Error(w, "Customer not found", http.StatusNotFound)
				return
			} else {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(authReq.Password))
	if err != nil {
		http.Error(w, "Incorrect password", http.StatusBadRequest)
		return
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": user.ID,
		"exp": time.Now().Add(time.Hour * 24 * 30).Unix(), // 24 hr expiration
	})

	// Sign and get the complete encoded token as a string using the secret
	tokenString, err := token.SignedString([]byte(os.Getenv("BCRYPT_KEY")))

	// Set the token as a cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "token",
		Value:    tokenString,
		Path:     "/",
		Expires:  time.Now().Add(time.Hour * 24 * 30),
		HttpOnly: true,
		Secure:   false,
	})

	w.Write([]byte("Authentication was successful."))
	w.WriteHeader(http.StatusOK)
}

func ListCustomers(w http.ResponseWriter, r *http.Request, rdb *redis.Client) {
	// Pagination: listcustomers?limit=10&offset=0
	customers := []models.Customer{}

	query := r.URL.Query()
	limitStr := query.Get("limit")
	offsetStr := query.Get("offset")

	// Provide defaults so no input required
	limit := 10
	offset := 0

	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil {
			limit = l
		}
	}
	if offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil {
			offset = o
		}
	}

	// Check Redis first
	ctx := context.Background()
	cacheKey := "customers:limit=" + strconv.Itoa(limit) + ":offset=" + strconv.Itoa(offset)
	cachedCustomers, err := cache.RedisClient.Client.Get(ctx, cacheKey).Result()
	if err == redis.Nil {
		log.Println("Cache miss. Retrieved from the database.")
		result := database.DB.Db.Unscoped().Limit(limit).Offset(offset).Find(&customers)
		if result.Error != nil && !errors.Is(result.Error, gorm.ErrRecordNotFound) {
			http.Error(w, result.Error.Error(), http.StatusInternalServerError)
			return
		}

		jsonResponse, err := json.Marshal(customers)
		if err != nil {
			http.Error(w, "Failed to serialize customers", http.StatusInternalServerError)
			return
		}

		// Cache customers
		if len(customers) > 0 {
			err = cache.RedisClient.Client.Set(ctx, cacheKey, jsonResponse, 10*time.Minute).Err()
			if err != nil {
				log.Printf("Redis SET error: %v", err)
				http.Error(w, "Failed to cache customers list", http.StatusInternalServerError)
				return
			}
		}

		w.WriteHeader(http.StatusOK)
		w.Write(jsonResponse)
	} else if err != nil {
		log.Printf("Redis GET error: %v", err)
		http.Error(w, "Failed to retrieve customers from cache", http.StatusInternalServerError)
	} else {
		log.Println("Retrieved from the cache.")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(cachedCustomers))
	}

}

func AddCustomer(w http.ResponseWriter, r *http.Request, rdb *redis.Client) {
	customer := new(models.Customer)

	err := json.NewDecoder(r.Body).Decode(&customer)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if customer.Email == "" {
		http.Error(w, "Missing customer email", http.StatusBadRequest)
		return
	}
	err = database.DB.Db.Create(&customer).Error
	if err != nil {
		http.Error(w, "Failed to add customer to the database", http.StatusInternalServerError)
		return
	}

	customerJSON, err := json.Marshal(customer)
	if err != nil {
		http.Error(w, "Failed to serialize customer data", http.StatusInternalServerError)
		return
	}

	ctx := context.Background()
	err = cache.RedisClient.Client.Set(ctx, "customer:"+customer.Email, customerJSON, 10*time.Minute).Err()
	if err != nil {
		log.Printf("Redis SET error: %v", err)
		http.Error(w, "Failed to add customer to the cache", http.StatusInternalServerError)
		return
	}

	err = cache.RedisClient.Client.Del(ctx, "customers:limit=10:offset=0").Err()
	if err != nil {
		log.Printf("Redis DEL error: %v", err)
		http.Error(w, "Failed to invalidate cache", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusAccepted)
	w.Write([]byte("Customer added successfully."))
}

func DeleteCustomer(w http.ResponseWriter, r *http.Request) {
	customer := new(models.Customer)

	err := json.NewDecoder(r.Body).Decode(&customer)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if customer.Email == "" {
		http.Error(w, "Missing customer email", http.StatusBadRequest)
		return
	} else {
		err = database.DB.Db.Unscoped().Where("email = ?", customer.Email).First(&customer).Error
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				http.Error(w, "Customer not found", http.StatusNotFound)
				return
			} else {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}
	}

	err = database.DB.Db.Unscoped().Delete(&customer).Error
	if err != nil {
		http.Error(w, "Failed to delete customer from database", http.StatusInternalServerError)
		return
	}

	ctx := context.Background()
	err = cache.RedisClient.Client.Del(ctx, "customer:"+customer.Email).Err()
	if err != nil {
		log.Printf("Redis DEL error: %v", err)
		http.Error(w, "Failed to delete the customer from the cache", http.StatusInternalServerError)
		return
	}
	// This is using cacheKey from ListCustomers but in strict string format.
	err = cache.RedisClient.Client.Del(ctx, "customers:limit=10:offset=0").Err()
	if err != nil {
		log.Printf("Redis DEL error: %v", err)
		http.Error(w, "Failed to invalidate cache", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Customer deleted successfully."))

}

func UpdateCustomer(w http.ResponseWriter, r *http.Request, rdb *redis.Client) {
	// Representation of the updated info
	var updatedinfo models.Customer

	// Representation of what's currently in the DB
	customer := new(models.Customer)

	err := json.NewDecoder(r.Body).Decode(&updatedinfo)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Must provide valid email; cannot update email via curl
	if updatedinfo.Email == "" {
		http.Error(w, "Missing customer email", http.StatusBadRequest)
		return
	}

	err = database.DB.Db.Where("email = ?", updatedinfo.Email).First(&customer).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			http.Error(w, "Customer not found", http.StatusNotFound)
			return
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	// Checking for empty fields to allow updating individual field without resetting the others
	if updatedinfo.Name != "" {
		customer.Name = updatedinfo.Name
	}
	if updatedinfo.Address != "" {
		customer.Address = updatedinfo.Address
	}
	if updatedinfo.Number != 0 {
		customer.Number = updatedinfo.Number
	}

	err = database.DB.Db.Save(&customer).Error
	if err != nil {
		http.Error(w, "Failed to update customer in database", http.StatusInternalServerError)
		return
	}

	// Update cache
	customerJSON, err := json.Marshal(customer)
	if err != nil {
		http.Error(w, "Failed to serialize customer data", http.StatusInternalServerError)
		return
	}

	ctx := context.Background()
	err = cache.RedisClient.Client.Set(ctx, "customer:"+customer.Email, customerJSON, 10*time.Minute).Err()
	if err != nil {
		log.Printf("Redis SET error: %v", err)
		http.Error(w, "Failed to update customer to the cache", http.StatusInternalServerError)
		return
	}
	// This is using cacheKey from ListCustomers but in strict string format.
	err = cache.RedisClient.Client.Del(ctx, "customers:limit=10:offset=0").Err()
	if err != nil {
		log.Printf("Redis DEL error: %v", err)
		http.Error(w, "Failed to invalidate cache", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Customer's information updated successfully."))
}
