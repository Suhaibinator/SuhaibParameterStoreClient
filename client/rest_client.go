package client

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
)

type APIClient struct {
	BaseURL                string
	AuthenticationPassword string
}

func NewAPIClient(baseURL, authenticationPassword string) *APIClient {
	return &APIClient{
		BaseURL:                baseURL,
		AuthenticationPassword: authenticationPassword,
	}
}

func (client *APIClient) Store(key, value string) error {
	url := fmt.Sprintf("%s/store", client.BaseURL)
	data := map[string]string{
		"key":   key,
		"value": value,
	}
	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", client.AuthenticationPassword)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusCreated {
		return errors.New("failed to create resource")

	}
	return nil
}

func (client *APIClient) Retrieve(key string) (string, error) {
	url := fmt.Sprintf("%s/retrieve?key=%s", client.BaseURL, key)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", client.AuthenticationPassword)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", errors.New("failed to retrieve data")
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var result map[string]string
	err = json.Unmarshal(body, &result)
	if err != nil {
		return "", err
	}

	return result["value"], nil
}

func RestSimpleRetrieve(ServerAddress string, AuthenticationPassword string, key string) (val string, err error) {
	client := NewAPIClient(ServerAddress, AuthenticationPassword)
	value, err := client.Retrieve(key)
	if err != nil {
		log.Println("Error retrieving data:", err)
	} else {
		log.Println("Retrieved value:", value)
	}
	return value, err
}

func RestSimpleStore(ServerAddress string, AuthenticationPassword string, key string, value string) (err error) {
	client := NewAPIClient(ServerAddress, AuthenticationPassword)
	err = client.Store(key, value)
	if err != nil {
		log.Println("Error storing data:", err)
	} else {
		log.Println("Data stored successfully")
	}
	return err
}
