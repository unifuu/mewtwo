package handler

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"order-service/model"
	"order-service/repository"

	"github.com/gin-gonic/gin"
	"github.com/streadway/amqp"
)

type OrderHandler struct {
	repo       repository.OrderRepository
	rabbitConn *amqp.Connection
}

func NewOrderHandler(repo repository.OrderRepository, rabbitConn *amqp.Connection) *OrderHandler {
	return &OrderHandler{
		repo:       repo,
		rabbitConn: rabbitConn,
	}
}

func (h *OrderHandler) CreateOrder(c *gin.Context) {
	var order model.Order
	if err := c.ShouldBindJSON(&order); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Set default status
	order.Status = "pending"

	if err := h.repo.Create(&order); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Publish message to RabbitMQ
	if err := h.publishOrderCreatedEvent(order); err != nil {
		log.Printf("Failed to publish order created event: %v", err)
	}

	c.JSON(http.StatusCreated, order)
}

func (h *OrderHandler) GetOrder(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	order, err := h.repo.FindByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Order not found"})
		return
	}

	c.JSON(http.StatusOK, order)
}

func (h *OrderHandler) GetOrdersByUserID(c *gin.Context) {
	userID, err := strconv.ParseUint(c.Param("user_id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	orders, err := h.repo.FindByUserID(uint(userID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, orders)
}

// Publish order created event to RabbitMQ
func (h *OrderHandler) publishOrderCreatedEvent(order model.Order) error {
	ch, err := h.rabbitConn.Channel()
	if err != nil {
		return err
	}
	defer ch.Close()

	// Declare exchange
	err = ch.ExchangeDeclare(
		"orders", // name
		"topic",  // type
		true,     // durable
		false,    // auto-deleted
		false,    // internal
		false,    // no-wait
		nil,      // arguments
	)
	if err != nil {
		return err
	}

	// Marshal order to JSON
	body, err := json.Marshal(order)
	if err != nil {
		return err
	}

	// Publish message
	err = ch.Publish(
		"orders",        // exchange
		"order.created", // routing key
		false,           // mandatory
		false,           // immediate
		amqp.Publishing{
			ContentType: "application/json",
			Body:        body,
		},
	)

	return err
}
