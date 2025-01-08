package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

const (
	apiScope = "https://www.googleapis.com/auth/chronicle-backstory"
	apiURL   = "https://backstory.googleapis.com/v1/dataTaps"
)

func main() {
	ctx := context.Background()

	defaultUsage := flag.Usage
	flag.Usage = func() {
		defaultUsage()

		fmt.Println(`
Commands:
  create  Create a new datatap
  list    List all datataps
  delete  Delete a datatap by topic ID`)
	}

	credentialsFile := flag.String("credentials", "", "Path to the credentials file")
	flag.Parse()

	var client *http.Client

	if *credentialsFile != "" {
		credentials, err := os.ReadFile(*credentialsFile)
		if err != nil {
			fmt.Println("Error reading credentials file:", err)
			return
		}

		creds, err := google.CredentialsFromJSON(ctx, credentials, apiScope)
		if err != nil {
			fmt.Println("Error loading credentials from file:", err)
			return
		}

		client = oauth2.NewClient(ctx, creds.TokenSource)
		if err != nil {
			fmt.Println("Error creating OAuth2 client:", err)
			return
		}
	} else {
		client = http.DefaultClient
	}

	// run command

	command := ""
	args := flag.Args()
	if len(args) >= 1 {
		command = args[0]
	}
	var resp *http.Response

	var err error
	switch command {
	case "create":
		resp, err = createCommand(client)
	case "list":
		resp, err = listCommand(client)
	case "delete":
		resp, err = deleteCommand(client)
	default:
		fmt.Println("Unknown command:\n", command)
		flag.Usage()
		os.Exit(1)
	}

	// handle response

	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Error reading response: %v\n", err)
		os.Exit(1)
	}

	if resp.StatusCode != http.StatusOK {
		fmt.Printf("Error: received status code %d\nResponse: %s\n", resp.StatusCode, body)
		os.Exit(1)
	}

	fmt.Printf("OK! Response: %s\n", body)
}

// createCommand creates a new datatap
func createCommand(client *http.Client) (*http.Response, error) {
	createCmd := flag.NewFlagSet("create", flag.ExitOnError)
	displayName := createCmd.String("displayName", "", "Display name for the datatap")
	topic := createCmd.String("topic", "", "Cloud Pub/Sub topic")
	serializationFormat := createCmd.String("format", "JSON", "Serialization format")

	if err := createCmd.Parse(flag.Args()[1:]); err != nil {
		return nil, fmt.Errorf("error parsing flags: %v", err)
	}

	if *displayName == "" || *topic == "" {
		createCmd.Usage()
		return nil, fmt.Errorf("both displayName and topic are required")
	}

	// create request
	var req struct {
		DisplayName     string `json:"displayName"`
		CloudPubsubSink struct {
			Topic string `json:"topic"`
		} `json:"cloudPubsubSink"`
		Filter              string `json:"filter"`
		SerializationFormat string `json:"serializationFormat"`
	}
	req.DisplayName = *displayName
	req.CloudPubsubSink.Topic = *topic
	req.Filter = "ALL_UDM_EVENTS"
	req.SerializationFormat = *serializationFormat

	reqBody, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("error marshaling request: %v", err)
	}

	resp, err := client.Post(apiURL, "application/json", bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, fmt.Errorf("error making request: %v", err)
	}

	return resp, nil
}

// listCommand lists all datataps
func listCommand(client *http.Client) (*http.Response, error) {
	resp, err := client.Get(apiURL)
	if err != nil {
		return nil, fmt.Errorf("error making request: %v", err)
	}

	return resp, nil
}

// deleteCommand deletes a datatap by topic ID
func deleteCommand(client *http.Client) (*http.Response, error) {
	deleteCmd := flag.NewFlagSet("delete", flag.ExitOnError)
	deleteCmd.Usage = func() {
		fmt.Println("Usage: delete <tapID>")
	}

	if err := deleteCmd.Parse(flag.Args()[1:]); err != nil {
		return nil, fmt.Errorf("error parsing flags: %v", err)
	}

	if deleteCmd.NArg() < 1 {
		deleteCmd.Usage()
		return nil, fmt.Errorf("tapID is required")
	}

	tapID := deleteCmd.Arg(0)
	url := fmt.Sprintf("%s/%s", apiURL, tapID)
	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %v", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error making request: %v", err)
	}

	return resp, nil
}
