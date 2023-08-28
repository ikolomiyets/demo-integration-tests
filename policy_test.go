package main

import (
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"io"
	"net/http"
	"testing"
	"time"
)

func TestGetPolicies(t *testing.T) {
	resp, err := http.Get(policyEndpoint + "/policies")
	if err != nil {
		logger.Fatalf("cannot access /policies endpoint: %v", err)
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

	var content PoliciesResponse
	err = json.Unmarshal(body, &content)

	if !assert.NotNil(t, content) {
		return
	}

	if !assert.NotNil(t, content.Policies) {
		return
	}

	assert.Equal(t, 3, len(content.Policies))
	assert.Equal(t, 1, content.First)
	assert.Equal(t, -1, content.Next)
	policy := content.Policies[0]
	location := policy.StartDate.Location()
	assert.Equal(t, "IE-8A-219276", policy.PolicyNumber)
	assert.Equal(t, time.Date(2017, time.May, 1, 0, 0, 0, 0, location), policy.StartDate)
	assert.Equal(t, 1, len(policy.Insured))
	assert.Equal(t, 2, policy.Insured[0])
	policy = content.Policies[1]
	assert.Equal(t, "IE-8A-210117", policy.PolicyNumber)
	assert.Equal(t, time.Date(2017, time.April, 17, 0, 0, 0, 0, location), policy.StartDate)
	assert.Equal(t, 2, len(policy.Insured))
	assert.Equal(t, 1, policy.Insured[0])
	assert.Equal(t, 4, policy.Insured[1])
	policy = content.Policies[2]
	assert.Equal(t, "IE-8C-001729", policy.PolicyNumber)
	assert.Equal(t, time.Date(2017, time.October, 11, 0, 0, 0, 0, location), policy.StartDate)
	assert.Equal(t, 1, len(policy.Insured))
	assert.Equal(t, 16, policy.Insured[0])
}
