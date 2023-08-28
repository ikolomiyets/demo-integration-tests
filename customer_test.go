package main

import (
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"io"
	"net/http"
	"testing"
)

func TestGetCustomersAll(t *testing.T) {
	resp, err := http.Get(customerEndpoint + "/customers")
	if err != nil {
		logger.Fatalf("cannot access /customers endpoint: %v", err)
		return
	}

	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("Expected status code %d. Got %d.", http.StatusNotFound, resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.Fatalf("cannot read response body: %v\n", err)
		return
	}

	var content Error
	err = json.Unmarshal(body, &content)

	if !assert.NotNil(t, content) {
		return
	}

	assert.Equal(t, "error", content.Message)
}

func TestGetCustomersNotFound(t *testing.T) {
	resp, err := http.Get(customerEndpoint + "/customers/3")
	if err != nil {
		logger.Fatalf("cannot access /customers endpoint: %v", err)
		return
	}

	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("Expected status code %d. Got %d.", http.StatusNotFound, resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.Fatalf("cannot read response body: %v\n", err)
		return
	}

	var content Error
	err = json.Unmarshal(body, &content)

	if !assert.NotNil(t, content) {
		return
	}

	assert.Equal(t, "", content.Message)
}

func TestGetCustomers(t *testing.T) {
	type TestCase struct {
		Name       string
		CustomerID string
		FirstName  string
		LastName   string
		Address    string
	}

	tests := []TestCase{
		{
			Name:       "customer 1",
			CustomerID: "1",
			FirstName:  "Sarah",
			LastName:   "Brennan",
			Address:    "Croom, Co. Limerick",
		},
		{
			Name:       "customer 2",
			CustomerID: "2",
			FirstName:  "Eva",
			LastName:   "Olson",
			Address:    "3, Patrick st, Limerick",
		},
		{
			Name:       "customer 4",
			CustomerID: "4",
			FirstName:  "James",
			LastName:   "Brennan",
			Address:    "Croom, Co. Limerick",
		},
		{
			Name:       "customer 16",
			CustomerID: "16",
			FirstName:  "Dermot",
			LastName:   "Finnegan",
			Address:    "Anacotty, Co. Limerick",
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			resp, err := http.Get(customerEndpoint + "/customers/" + test.CustomerID)
			if err != nil {
				logger.Fatalf("cannot access /customers endpoint: %v", err)
				return
			}

			if resp.StatusCode != http.StatusOK {
				t.Errorf("Expected status code %d. Got %d.", http.StatusOK, resp.StatusCode)
			}

			body, err := io.ReadAll(resp.Body)
			if err != nil {
				logger.Fatalf("cannot read response body: %v\n", err)
				return
			}

			var content Customer
			err = json.Unmarshal(body, &content)

			if !assert.NotNil(t, content) {
				return
			}

			assert.Equal(t, test.FirstName, content.FirstName)
			assert.Equal(t, test.LastName, content.LastName)
			assert.Equal(t, test.Address, content.Address)
		})
	}
}
