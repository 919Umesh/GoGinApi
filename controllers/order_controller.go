package controllers

import (
	"fmt"
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

	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
		return
	}

	var userIDUint uint
	switch v := userID.(type) {
	case float64:
		userIDUint = uint(v)
	case int:
		userIDUint = uint(v)
	case int64:
		userIDUint = uint(v)
	case uint:
		userIDUint = v
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid user ID type"})
		return
	}

	fmt.Println("-----------Begin Transaction---------------")
	tx, err := config.DB.Begin()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	var totalAmount float64
	var orderItems []models.OrderItem

	for _, item := range req.Items {
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

		if product.Quantity < item.Quantity {
			tx.Rollback()
			c.JSON(http.StatusBadRequest, gin.H{
				"error": fmt.Sprintf("insufficient stock for product %s (available: %d, requested: %d)",
					product.Name, product.Quantity, item.Quantity),
			})
			return
		}

		itemTotal := product.Price * float64(item.Quantity)
		totalAmount += itemTotal

		orderItems = append(orderItems, models.OrderItem{
			ProductID:  product.ID,
			Quantity:   item.Quantity,
			UnitPrice:  product.Price,
			TotalPrice: itemTotal,
		})
	}

	result, err := tx.Exec(`
		INSERT INTO orders (user_id, total_amount) 
		VALUES (?, ?)`,
		userIDUint, totalAmount,
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

	for _, item := range orderItems {
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

	if err := tx.Commit(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	order := models.Order{
		ID:          uint(orderID),
		UserID:      userIDUint,
		TotalAmount: totalAmount,
		Status:      "pending",
		OrderItems:  orderItems,
	}

	c.JSON(http.StatusCreated, order)
}

func GetUserOrders(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
		return
	}

	var userIDUint uint
	switch v := userID.(type) {
	case float64:
		userIDUint = uint(v)
	case int:
		userIDUint = uint(v)
	case int64:
		userIDUint = uint(v)
	case uint:
		userIDUint = v
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid user ID type"})
		return
	}

	rows, err := config.DB.Query(`
		SELECT id, user_id, total_amount, status, created_at 
		FROM orders 
		WHERE user_id = ? 
		ORDER BY created_at DESC`,
		userIDUint,
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
			var product models.Product
			if err := itemRows.Scan(
				&item.ID,
				&item.ProductID,
				&item.Quantity,
				&item.UnitPrice,
				&item.TotalPrice,
				&product.Name,
				&product.Image,
				&product.SalesRate,
				&product.PurchaseRate,
			); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			item.Product = product
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

	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
		return
	}

	var userIDUint uint
	switch v := userID.(type) {
	case float64:
		userIDUint = uint(v)
	case int:
		userIDUint = uint(v)
	case int64:
		userIDUint = uint(v)
	case uint:
		userIDUint = v
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid user ID type"})
		return
	}

	var order models.Order
	err = config.DB.QueryRow(`
		SELECT id, user_id, total_amount, status, created_at 
		FROM orders 
		WHERE id = ? AND user_id = ?`,
		orderID, userIDUint,
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
		var product models.Product
		if err := itemRows.Scan(
			&item.ID,
			&item.ProductID,
			&item.Quantity,
			&item.UnitPrice,
			&item.TotalPrice,
			&product.Name,
			&product.Image,
			&product.SalesRate,
			&product.PurchaseRate,
		); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		item.Product = product
		orderItems = append(orderItems, item)
	}
	order.OrderItems = orderItems

	c.JSON(http.StatusOK, order)
}
