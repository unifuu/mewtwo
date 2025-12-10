package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"user-service/handler"
	"user-service/model"
	"user-service/repository"

	"github.com/gin-gonic/gin"
	"github.com/hashicorp/consul/api"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var (
	consulClient *api.Client
	serviceID    = "user-service-1"
	serviceName  = "user-service"
	servicePort  = 8081
	serviceHost  string
)

func main() {
	// Get database config from environment or use defaults
	dbHost := getEnv("DB_HOST", "localhost")
	dbPort := getEnv("DB_PORT", "5432")
	dbUser := getEnv("DB_USER", "postgres")
	dbPassword := getEnv("DB_PASSWORD", "postgres123")
	dbName := getEnv("DB_NAME", "user_db")

	// Get service config from environment
	serviceName = getEnv("SERVICE_NAME", serviceName)
	servicePort = getPortFromEnv("SERVICE_PORT", servicePort)
	serviceHost = getEnv("SERVICE_HOST", serviceName) // Default to service name for Docker networking

	// Connect to database
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
		dbHost, dbUser, dbPassword, dbName, dbPort)
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// Auto migrate
	db.AutoMigrate(&model.User{})

	// Initialize repository and handler
	userRepo := repository.NewUserRepository(db)
	userHandler := handler.NewUserHandler(userRepo)

	// Register service to Consul
	if err := registerService(); err != nil {
		log.Fatal("Failed to register service:", err)
	}

	// Setup router
	r := gin.Default()

	// Health check endpoint
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "healthy"})
	})

	// User endpoints
	r.POST("/users", userHandler.CreateUser)
	r.GET("/users/:id", userHandler.GetUser)
	r.GET("/users", userHandler.ListUsers)
	r.GET("/users/:id/orders", userHandler.GetUserOrders)

	// Graceful shutdown
	go func() {
		if err := r.Run(fmt.Sprintf(":%d", servicePort)); err != nil {
			log.Fatal("Failed to start server:", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	// Deregister service
	deregisterService()
	log.Println("User service stopped")
}

func registerService() error {
	// Get Consul config from environment
	consulHost := getEnv("CONSUL_HOST", "localhost")
	consulPort := getEnv("CONSUL_PORT", "8500")
	consulAddress := fmt.Sprintf("%s:%s", consulHost, consulPort)

	// Create Consul client
	config := api.DefaultConfig()
	config.Address = consulAddress

	var err error
	consulClient, err = api.NewClient(config)
	if err != nil {
		return err
	}

	// Register service
	registration := &api.AgentServiceRegistration{
		ID:      serviceID,
		Name:    serviceName,
		Port:    servicePort,
		Address: serviceHost,
		Check: &api.AgentServiceCheck{
			HTTP:                           fmt.Sprintf("http://%s:%d/health", serviceHost, servicePort),
			Interval:                       "10s",
			Timeout:                        "3s",
			DeregisterCriticalServiceAfter: "30s",
		},
	}

	return consulClient.Agent().ServiceRegister(registration)
}

func deregisterService() {
	if consulClient != nil {
		consulClient.Agent().ServiceDeregister(serviceID)
	}
}

func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

func getPortFromEnv(key string, defaultValue int) int {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	var port int
	fmt.Sscanf(value, "%d", &port)
	if port == 0 {
		return defaultValue
	}
	return port
}
