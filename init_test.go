package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/testcontainers/testcontainers-go"
	"io"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"
)

var (
	policyEndpoint   string
	customerEndpoint string
	frontendEndpoint string
	readToken        string
	logger           *log.Logger
	client           *http.Client
)

func TestMain(m *testing.M) {
	logger = log.New(os.Stdout, "test: ", log.LstdFlags)

	readToken = os.Getenv("ARTIFACTZ_TOKEN")

	isLocal := false
	localString := os.Getenv("TEST_LOCAL")
	if localString != "" {
		b, err := strconv.ParseBool(localString)
		if err == nil {
			isLocal = b
		}
	}

	showOutput := false
	showContainerOutputString := os.Getenv("SHOW_CONTAINERS_OUTPUT")
	if showContainerOutputString != "" {
		b, err := strconv.ParseBool(showContainerOutputString)
		if err == nil {
			showOutput = b
		}
	}

	tr := &http.Transport{
		MaxIdleConns:       10,
		IdleConnTimeout:    30 * time.Second,
		DisableCompression: true,
	}
	client = &http.Client{Transport: tr}

	versions, err := getImageVersions(client, readToken)
	if err != nil {
		logger.Fatal(err)
	}

	policyImage := "ikolomiyets/demo-policy:" + versions["demo-policy"]
	logger.Printf("pulling policy image: %v", policyImage)
	customerImage := "ikolomiyets/demo-customers:" + versions["demo-customers"]
	frontendImage := "ikolomiyets/demo-frontend:" + versions["demo-frontend"]

	if isLocal {
		logger.Println("Local integration test: using local images")
		policyImage = "demo-policy"
		customerImage = "demo-customer"
		frontendImage = "demo-frontend"
	}

	ctx := context.Background()

	networkName := "test-network-" + randomString()

	networkRequest := testcontainers.NetworkRequest{
		Driver:     "bridge",
		Name:       networkName,
		Attachable: true,
	}

	env := make(map[string]string)

	policyContainerName := "demo-policy-" + randomString()

	env = make(map[string]string)
	env["DEBUG"] = "true"

	//WaitingFor:      wait.ForListeningPort("5000/tcp"),
	policyRequest := testcontainers.ContainerRequest{
		Image:           policyImage,
		Name:            policyContainerName,
		Hostname:        "policies",
		ExposedPorts:    []string{"8080/tcp"},
		Env:             env,
		AutoRemove:      true,
		Networks:        []string{networkName},
		AlwaysPullImage: !isLocal,
	}

	gcr := testcontainers.GenericContainerRequest{
		ContainerRequest: policyRequest,
		Started:          true,
	}

	provider, err := gcr.ProviderType.GetProvider()
	if err != nil {
		logger.Fatalln("cannot get provider")
	}

	net, err := provider.CreateNetwork(ctx, networkRequest)
	if err != nil {
		logger.Fatalln("cannot create network")
	}

	defer net.Remove(ctx)

	policy, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: policyRequest,
		Started:          true,
	})
	if err != nil {
		logger.Fatalln(err)
		return
	}

	defer policy.Terminate(ctx)

	err = policy.StartLogProducer(ctx)
	if err != nil {
		logger.Fatalln(err)
		return
	}

	defer policy.StopLogProducer()

	policyLogConsumer := NewLogConsoleConsumer("policy", logger, "Started PolicyDemoApplication in", showOutput)
	policy.FollowOutput(policyLogConsumer)

	policyEndpoint, err = policy.PortEndpoint(ctx, "8080/tcp", "http")
	if err != nil {
		logger.Fatalln(err)
		return
	}

	env = make(map[string]string)
	env["DEBUG"] = "true"

	customerContainerName := "demo-customer-" + randomString()
	customerRequest := testcontainers.ContainerRequest{
		Image:           customerImage,
		Name:            customerContainerName,
		Hostname:        "customers",
		ExposedPorts:    []string{"3000/tcp"},
		Env:             env,
		AutoRemove:      true,
		Networks:        []string{networkName},
		AlwaysPullImage: !isLocal,
	}

	customer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: customerRequest,
		Started:          true,
	})
	if err != nil {
		logger.Fatalln(err)
		return
	}

	defer customer.Terminate(ctx)

	err = customer.StartLogProducer(ctx)
	if err != nil {
		logger.Fatalln(err)
		return
	}

	defer customer.StopLogProducer()

	customerLogConsumer := NewLogConsoleConsumer("customer", logger, "", showOutput)
	customer.FollowOutput(customerLogConsumer)

	customerEndpoint, err = customer.Endpoint(ctx, "http")
	if err != nil {
		logger.Fatalln(err)
		return
	}

	frontendContainerName := "frontend-" + randomString()
	env = make(map[string]string)
	env["DEBUG"] = "true"

	frontendRequest := testcontainers.ContainerRequest{
		Image:           frontendImage,
		Name:            frontendContainerName,
		ExposedPorts:    []string{"80/tcp"},
		Env:             env,
		AutoRemove:      true,
		Networks:        []string{networkName},
		AlwaysPullImage: !isLocal,
	}
	frontend, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: frontendRequest,
		Started:          true,
	})
	if err != nil {
		logger.Fatalln(err)
		return
	}

	defer frontend.Terminate(ctx)

	err = frontend.StartLogProducer(ctx)
	if err != nil {
		logger.Fatalln(err)
		return
	}

	defer frontend.StopLogProducer()

	frontendLogConsumer := NewLogConsoleConsumer("frontend", logger, "", showOutput)
	frontend.FollowOutput(frontendLogConsumer)

	frontendEndpoint, err = frontend.Endpoint(ctx, "http")
	if err != nil {
		logger.Fatalln(err)
		return
	}

	err = policyLogConsumer.WaitForContainerToStart()
	if err != nil {
		logger.Fatalln(err)
		return
	}

	logger.Println("containers are ready")

	retval := m.Run()

	os.Exit(retval)
}

func getImageVersions(client *http.Client, token string) (map[string]string, error) {
	request, err := http.NewRequest("GET", "https://artifactor.artifactz.io/stages/Integration Test/list?artifact=demo-policy&artifact=demo-customers&artifact=demo-frontend", nil)
	if err != nil {
		logger.Fatalf("failed to build new request to https://artifactor.artifactz.io: %v", err)
	}

	request.Header.Add("Authorization", "Bearer "+token)
	request.Header.Add("Accept", "application/json")
	request.Header.Add("X-ClientId", "Integration Test")

	resp, err := client.Do(request)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Error running request: %v", err))
	}
	defer resp.Body.Close()

	var (
		body      []byte
		artifacts StageArtifacts
		result    map[string]string
	)

	body, err = io.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("cannot read response body: %v", err))
	}

	err = json.Unmarshal(body, &artifacts)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("cannot handle response: %v", err))
	}

	result = make(map[string]string)
	for _, version := range artifacts.Artifacts {
		result[version.ArtifactName] = version.Version
	}

	return result, nil
}

func randomString() string {
	rand.Seed(time.Now().UnixNano())
	chars := []rune("ABCDEFGHIJKLMNOPQRSTUVWXYZ" +
		"abcdefghijklmnopqrstuvwxyz" +
		"0123456789")
	length := 8
	var b strings.Builder
	for i := 0; i < length; i++ {
		b.WriteRune(chars[rand.Intn(len(chars))])
	}
	return b.String()
}
