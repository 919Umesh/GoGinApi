package models

type Product struct {
	ID           uint    `json:"id"`
	Name         string  `json:"name"`
	Price        float64 `json:"price"`
	Quantity     int     `json:"quantity"`
	Image        string  `json:"image"`
	SalesRate    float64 `json:"sales_rate"`
	PurchaseRate float64 `json:"purchase_rate"`
}
