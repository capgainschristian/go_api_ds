package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/capgainschristian/go_api_ds/database"
	"github.com/capgainschristian/go_api_ds/models"
	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"
)

func HealthCheck(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)

	fmt.Fprint(w, "API is up and running.")
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
	cachedCustomers, err := rdb.Get(ctx, cacheKey).Result()
	if err == redis.Nil {
		log.Println("Retrieved from the database.")
		result := database.DB.Db.Limit(limit).Offset(offset).Find(&customers)
		if result.Error != nil {
			http.Error(w, result.Error.Error(), http.StatusInternalServerError)
			return
		}

		jsonResponse, err := json.Marshal(customers)
		if err != nil {
			http.Error(w, "Failed to marshal customers", http.StatusInternalServerError)
			return
		}

		// Cache customers
		err = rdb.Set(ctx, cacheKey, jsonResponse, 24*time.Hour).Err()
		if err != nil {
			log.Printf("Redis SET error: %v", err)
			http.Error(w, "Failed to cache customers list", http.StatusInternalServerError)
			return
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
	err = rdb.Set(ctx, "customer:"+strconv.Itoa(int(customer.ID)), customerJSON, 24*time.Hour).Err()
	if err != nil {
		log.Printf("Redis SET error: %v", err)
		http.Error(w, "Failed to add customer to the cache", http.StatusInternalServerError)
		return
	}

	err = rdb.Del(ctx, "customers:limit=10:offset=0").Err()
	if err != nil {
		log.Printf("Redis DEL error: %v", err) // Log the error for debugging
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
		err = database.DB.Db.Where("email = ?", customer.Email).First(&customer).Error
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

	err = database.DB.Db.Delete(&customer).Error
	if err != nil {
		http.Error(w, "Failed to delete customer from database", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Customer deleted successfully."))

}

func UpdateCustomer(w http.ResponseWriter, r *http.Request) {
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

	customer.Name = updatedinfo.Name
	customer.Email = updatedinfo.Email
	customer.Address = updatedinfo.Address
	customer.Number = updatedinfo.Number

	err = database.DB.Db.Save(&customer).Error
	if err != nil {
		http.Error(w, "Failed to update customer in database", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Customer's information updated successfully."))
}
