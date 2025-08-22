// Order Service Microservice
// Handles order creation, processing, and management

package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// Order represents an order in the system
type Order struct {
	ID          string    `json:"id"`
	UserID      string    `json:"user_id"`
	Items       []Item    `json:"items"`
	TotalAmount float64   `json:"total_amount"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// Item represents an item in an order
type Item struct {
	ID       string  `json:"id"`
	Name     string  `json:"name"`
	Price    float64 `json:"price"`
	Quantity int     `json:"quantity"`
}

// CreateOrderRequest represents the request to create an order
type CreateOrderRequest struct {
	UserID string `json:"user_id" binding:"required"`
	Items  []Item `json:"items" binding:"required"`
}

// In-memory storage (in production, use a real database)
var orders []Order

func main() {
	// Initialize Gin
	if os.Getenv("ENVIRONMENT") == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.Default()

	// Middleware
	r.Use(gin.Logger())
	r.Use(gin.Recovery())
	r.Use(corsMiddleware())

	// Initialize with sample data
	initSampleData()

	// Routes
	r.GET("/health", healthCheck)
	r.POST("/orders", createOrder)
	r.GET("/orders", getOrders)
	r.GET("/orders/:id", getOrder)
	r.PUT("/orders/:id/status", updateOrderStatus)
	r.DELETE("/orders/:id", deleteOrder)

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "3002"
	}

	log.Printf("🚀 Order Service running on port %s", port)
	log.Fatal(r.Run(":" + port))
}

func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

func initSampleData() {
	orders = []Order{
		{
			ID:     "order-1",
			UserID: "user-1",
			Items: []Item{
				{ID: "item-1", Name: "Laptop", Price: 999.99, Quantity: 1},
				{ID: "item-2", Name: "Mouse", Price: 29.99, Quantity: 2},
			},
			TotalAmount: 1059.97,
			Status:      "completed",
			CreatedAt:   time.Now().Add(-24 * time.Hour),
			UpdatedAt:   time.Now().Add(-23 * time.Hour),
		},
	}
}

func healthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":    "healthy",
		"service":   "order-service",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
		"version":   "1.0.0",
		"orders_count": len(orders),
	})
}

func createOrder(c *gin.Context) {
	var req CreateOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Calculate total amount
	var totalAmount float64
	for _, item := range req.Items {
		totalAmount += item.Price * float64(item.Quantity)
	}

	// Create new order
	order := Order{
		ID:          uuid.New().String(),
		UserID:      req.UserID,
		Items:       req.Items,
		TotalAmount: totalAmount,
		Status:      "pending",
		CreatedAt:   time.Now().UTC(),
		UpdatedAt:   time.Now().UTC(),
	}

	orders = append(orders, order)

	log.Printf("Created order %s for user %s with total amount %.2f", order.ID, order.UserID, order.TotalAmount)

	c.JSON(http.StatusCreated, order)
}

func getOrders(c *gin.Context) {
	userID := c.Query("user_id")
	status := c.Query("status")
	limit := c.DefaultQuery("limit", "10")

	limitInt, err := strconv.Atoi(limit)
	if err != nil {
		limitInt = 10
	}

	var filteredOrders []Order
	for _, order := range orders {
		if userID != "" && order.UserID != userID {
			continue
		}
		if status != "" && order.Status != status {
			continue
		}
		filteredOrders = append(filteredOrders, order)
		
		if len(filteredOrders) >= limitInt {
			break
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"orders": filteredOrders,
		"total":  len(filteredOrders),
	})
}

func getOrder(c *gin.Context) {
	orderID := c.Param("id")

	for _, order := range orders {
		if order.ID == orderID {
			c.JSON(http.StatusOK, order)
			return
		}
	}

	c.JSON(http.StatusNotFound, gin.H{"error": "Order not found"})
}

func updateOrderStatus(c *gin.Context) {
	orderID := c.Param("id")
	var req struct {
		Status string `json:"status" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	for i, order := range orders {
		if order.ID == orderID {
			orders[i].Status = req.Status
			orders[i].UpdatedAt = time.Now().UTC()
			
			log.Printf("Updated order %s status to %s", orderID, req.Status)
			
			c.JSON(http.StatusOK, orders[i])
			return
		}
	}

	c.JSON(http.StatusNotFound, gin.H{"error": "Order not found"})
}

func deleteOrder(c *gin.Context) {
	orderID := c.Param("id")

	for i, order := range orders {
		if order.ID == orderID {
			orders = append(orders[:i], orders[i+1:]...)
			
			log.Printf("Deleted order %s", orderID)
			
			c.JSON(http.StatusOK, gin.H{"message": "Order deleted successfully"})
			return
		}
	}

	c.JSON(http.StatusNotFound, gin.H{"error": "Order not found"})
}
