package client

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/hashicorp/consul/api"
)

type OrderClient struct {
	consulClient *api.Client
}

type Order struct {
	ID        uint    `json:"id"`
	UserID    uint    `json:"user_id"`
	ProductID uint    `json:"product_id"`
	Amount    float64 `json:"amount"`
}

func NewOrderClient() *OrderClient {
	// Get Consul config from environment
	consulHost := getEnv("CONSUL_HOST", "localhost")
	consulPort := getEnv("CONSUL_PORT", "8500")
	consulAddress := fmt.Sprintf("%s:%s", consulHost, consulPort)

	config := api.DefaultConfig()
	config.Address = consulAddress

	client, err := api.NewClient(config)
	if err != nil {
		panic(err)
	}

	return &OrderClient{consulClient: client}
}

func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

// Service discovery: get order service address from Consul
func (c *OrderClient) getServiceAddress() (string, error) {
	services, _, err := c.consulClient.Health().Service("order-service", "", true, nil)
	if err != nil {
		return "", err
	}

	if len(services) == 0 {
		return "", fmt.Errorf("no healthy order service found")
	}

	// Simple load balancing: return first healthy service
	service := services[0]
	address := fmt.Sprintf("http://%s:%d", service.Service.Address, service.Service.Port)

	return address, nil
}

func (c *OrderClient) GetOrdersByUserID(userID uint) ([]Order, error) {
	address, err := c.getServiceAddress()
	if err != nil {
		return nil, fmt.Errorf("failed to get order service address: %w", err)
	}

	url := fmt.Sprintf("%s/orders/user/%d", address, userID)
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to call order service: %w", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("order service returned status %d: %s", resp.StatusCode, string(body))
	}

	var orders []Order
	if err := json.Unmarshal(body, &orders); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return orders, nil
}
