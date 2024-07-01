package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os/exec"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/capgainschristian/go_api_ds/models"
)

func main() {
	// Seed the random generator for consistent results (optional)
	gofakeit.Seed(0)

	numDataPoints := 100

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

		// Need to capture address as string for CURL
		address := fmt.Sprintf("%s, %s, %s, %s",
			dataPoint["street"],
			dataPoint["city"],
			dataPoint["state"],
			dataPoint["zip"],
		)

		// Prepare customer data using the random dataPoints
		customer := models.Customer{
			Name:    dataPoint["name"].(string),
			Email:   dataPoint["email"].(string),
			Address: address,
			Number:  dataPoint["number"].(int),
		}

		// Send as JSON
		jsonData, err := json.Marshal(customer)
		if err != nil {
			log.Printf("Error marshaling JSON: %v\nData Point: %v\n", err, dataPoint)
		}

		curlCommand := exec.Command("curl",
			"--header", "Content-Type: application/json",
			"--request", "POST",
			"--data", string(jsonData),
			"http://localhost:3000/customercreation",
		)

		output, err := curlCommand.CombinedOutput()

		if err != nil {
			log.Printf("Error executing curl command: %v\nOutput: %s\nData Point: %v\n", err, output, dataPoint)
			continue
		}

		log.Printf("Customer added successfully: %s\n", customer.Name)

	}
}
