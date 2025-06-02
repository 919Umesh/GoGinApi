package models

import "time"

type Order struct {
	ID          uint        `json:"id"`
	UserID      uint        `json:"user_id"`
	TotalAmount float64     `json:"total_amount"`
	Status      string      `json:"status"`
	CreatedAt   time.Time   `json:"created_at"`
	OrderItems  []OrderItem `json:"order_items"`
}

type OrderItem struct {
	ID         uint    `json:"id"`
	OrderID    uint    `json:"order_id"`
	ProductID  uint    `json:"product_id"`
	Quantity   int     `json:"quantity"`
	UnitPrice  float64 `json:"unit_price"`
	TotalPrice float64 `json:"total_price"`
	Product    Product `json:"product"`
}

type CreateOrderRequest struct {
	Items []struct {
		ProductID uint `json:"product_id" binding:"required"`
		Quantity  int  `json:"quantity" binding:"required,min=1"`
	} `json:"items" binding:"required,min=1"`
}
