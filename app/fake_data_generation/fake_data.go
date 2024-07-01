package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os/exec"
	"sync"

	"github.com/brianvoe/gofakeit/v6"
)

type Customer struct {
	Name    string `json:"name"`
	Email   string `json:"email"`
	Address string `json:"address"`
	Number  int    `json:"number"`
}

func main() {
	gofakeit.Seed(0)

	numDataPoints := 100
	dataPoints := generateDataPoints(numDataPoints)

	// Create a channel to communicate results
	resultChan := make(chan error, numDataPoints)

	// Use a wait group to wait for all goroutines to finish
	var wg sync.WaitGroup
	wg.Add(numDataPoints)

	// Launch goroutines to handle customer creation concurrently
	for _, dataPoint := range dataPoints {
		go func(dp map[string]interface{}) {
			defer wg.Done()

			address := fmt.Sprintf("%s, %s, %s, %s",
				dp["street"],
				dp["city"],
				dp["state"],
				dp["zip"],
			)

			customer := Customer{
				Name:    dp["name"].(string),
				Email:   dp["email"].(string),
				Address: address,
				Number:  dp["number"].(int),
			}

			jsonData, err := json.Marshal(customer)
			if err != nil {
				log.Printf("Error marshaling JSON: %v\nData Point: %v\n", err, dp)
				resultChan <- err
				return
			}

			curlCommand := exec.Command("curl",
				"--header", "Content-Type: application/json",
				"--request", "POST",
				"--data", string(jsonData),
				"http://localhost:3000/customercreation",
			)

			output, err := curlCommand.CombinedOutput()
			if err != nil {
				log.Printf("Error executing curl command: %v\nOutput: %s\nData Point: %v\n", err, output, dp)
				resultChan <- err
				return
			}

			log.Printf("Customer added successfully: %s\n", customer.Name)
			resultChan <- nil // Successfully added customer
		}(dataPoint)
	}

	// Close the result channel after all goroutines finish
	go func() {
		wg.Wait()
		close(resultChan)
	}()

	// Process results from result channel
	for result := range resultChan {
		if result != nil {
			log.Printf("Error adding customer: %v\n", result)
		}
	}
}

// Function to generate random data points
func generateDataPoints(numDataPoints int) []map[string]interface{} {
	dataPoints := make([]map[string]interface{}, numDataPoints)

	for i := 0; i < numDataPoints; i++ {
		dataPoint := make(map[string]interface{})

		addressData := gofakeit.Address()
		dataPoint["name"] = gofakeit.Name()
		dataPoint["email"] = gofakeit.Email()
		dataPoint["street"] = addressData.Street
		dataPoint["city"] = addressData.City
		dataPoint["state"] = addressData.State
		dataPoint["zip"] = addressData.Zip
		dataPoint["number"] = gofakeit.Number(1, 100)

		dataPoints[i] = dataPoint
	}

	return dataPoints
}
