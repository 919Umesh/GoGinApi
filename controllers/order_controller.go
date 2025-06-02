package controllers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/umesh/ginapi/config"
	"github.com/umesh/ginapi/models"
)

func CreateOrder(c *gin.Context) {
	var req models.CreateOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get user ID from context (set by AuthMiddleware)
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
		return
	}

	// Begin transaction
	tx, err := config.DB.Begin()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Calculate total amount and prepare order items
	var totalAmount float64
	var orderItems []models.OrderItem

	for _, item := range req.Items {
		// Get product details
		var product models.Product
		err := tx.QueryRow(`
			SELECT id, name, price, quantity 
			FROM products 
			WHERE id = ? FOR UPDATE`,
			item.ProductID,
		).Scan(&product.ID, &product.Name, &product.Price, &product.Quantity)
		if err != nil {
			tx.Rollback()
			c.JSON(http.StatusBadRequest, gin.H{"error": "product not found"})
			return
		}

		// Check stock availability
		if product.Quantity < item.Quantity {
			tx.Rollback()
			c.JSON(http.StatusBadRequest, gin.H{"error": "insufficient stock for product " + product.Name})
			return
		}

		// Calculate item total
		itemTotal := product.Price * float64(item.Quantity)
		totalAmount += itemTotal

		// Prepare order item
		orderItems = append(orderItems, models.OrderItem{
			ProductID:  product.ID,
			Quantity:   item.Quantity,
			UnitPrice:  product.Price,
			TotalPrice: itemTotal,
		})
	}

	// Create order
	result, err := tx.Exec(`
		INSERT INTO orders (user_id, total_amount) 
		VALUES (?, ?)`,
		userID, totalAmount,
	)
	if err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	orderID, err := result.LastInsertId()
	if err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Create order items and update product quantities
	for _, item := range orderItems {
		// Insert order item
		_, err := tx.Exec(`
			INSERT INTO order_items (order_id, product_id, quantity, unit_price, total_price) 
			VALUES (?, ?, ?, ?, ?)`,
			orderID, item.ProductID, item.Quantity, item.UnitPrice, item.TotalPrice,
		)
		if err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// Update product quantity
		_, err = tx.Exec(`
			UPDATE products 
			SET quantity = quantity - ? 
			WHERE id = ?`,
			item.Quantity, item.ProductID,
		)
		if err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Return created order
	order := models.Order{
		ID:          uint(orderID),
		UserID:      userID.(uint),
		TotalAmount: totalAmount,
		Status:      "pending",
		OrderItems:  orderItems,
	}

	c.JSON(http.StatusCreated, order)
}

func GetUserOrders(c *gin.Context) {
	// Get user ID from context
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
		return
	}

	// Query orders
	rows, err := config.DB.Query(`
		SELECT id, user_id, total_amount, status, created_at 
		FROM orders 
		WHERE user_id = ? 
		ORDER BY created_at DESC`,
		userID,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var orders []models.Order
	for rows.Next() {
		var order models.Order
		if err := rows.Scan(
			&order.ID,
			&order.UserID,
			&order.TotalAmount,
			&order.Status,
			&order.CreatedAt,
		); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		orders = append(orders, order)
	}

	// Get order items for each order
	for i, order := range orders {
		itemRows, err := config.DB.Query(`
			SELECT oi.id, oi.product_id, oi.quantity, oi.unit_price, oi.total_price, 
			       p.name, p.image, p.sales_rate, p.purchase_rate
			FROM order_items oi
			JOIN products p ON oi.product_id = p.id
			WHERE oi.order_id = ?`,
			order.ID,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		defer itemRows.Close()

		var orderItems []models.OrderItem
		for itemRows.Next() {
			var item models.OrderItem
			if err := itemRows.Scan(
				&item.ID,
				&item.ProductID,
				&item.Quantity,
				&item.UnitPrice,
				&item.TotalPrice,
				&item.Product.Name,
				&item.Product.Image,
				&item.Product.SalesRate,
				&item.Product.PurchaseRate,
			); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			orderItems = append(orderItems, item)
		}
		orders[i].OrderItems = orderItems
	}

	c.JSON(http.StatusOK, orders)
}

func GetOrderByID(c *gin.Context) {
	orderID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid order ID"})
		return
	}

	// Get user ID from context
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
		return
	}

	// Query order
	var order models.Order
	err = config.DB.QueryRow(`
		SELECT id, user_id, total_amount, status, created_at 
		FROM orders 
		WHERE id = ? AND user_id = ?`,
		orderID, userID,
	).Scan(
		&order.ID,
		&order.UserID,
		&order.TotalAmount,
		&order.Status,
		&order.CreatedAt,
	)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "order not found"})
		return
	}

	// Get order items
	itemRows, err := config.DB.Query(`
		SELECT oi.id, oi.product_id, oi.quantity, oi.unit_price, oi.total_price, 
		       p.name, p.image, p.sales_rate, p.purchase_rate
		FROM order_items oi
		JOIN products p ON oi.product_id = p.id
		WHERE oi.order_id = ?`,
		order.ID,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer itemRows.Close()

	var orderItems []models.OrderItem
	for itemRows.Next() {
		var item models.OrderItem
		if err := itemRows.Scan(
			&item.ID,
			&item.ProductID,
			&item.Quantity,
			&item.UnitPrice,
			&item.TotalPrice,
			&item.Product.Name,
			&item.Product.Image,
			&item.Product.SalesRate,
			&item.Product.PurchaseRate,
		); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		orderItems = append(orderItems, item)
	}
	order.OrderItems = orderItems

	c.JSON(http.StatusOK, order)
}
