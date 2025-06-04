package controllers

import (
	"math"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/umesh/ginapi/config"
	"github.com/umesh/ginapi/models"
)

// Get All Products
func GetProducts(c *gin.Context) {
	page, err := strconv.Atoi(c.DefaultQuery("page", "1"))
	if err != nil || page < 1 {
		page = 1
	}

	limit, err := strconv.Atoi(c.DefaultQuery("limit", "10"))
	if err != nil || limit < 1 {
		limit = 10
	}

	offset := (page - 1) * limit

	query := `
        SELECT id, name, price, quantity, image, sales_rate, purchase_rate 
        FROM products 
        LIMIT ? OFFSET ?
    `
	rows, err := config.DB.Query(query, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var products []models.Product
	for rows.Next() {
		var product models.Product
		err := rows.Scan(
			&product.ID,
			&product.Name,
			&product.Price,
			&product.Quantity,
			&product.Image,
			&product.SalesRate,
			&product.PurchaseRate,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		products = append(products, product)
	}

	if err = rows.Err(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var total int
	err = config.DB.QueryRow("SELECT COUNT(*) FROM products").Scan(&total)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": products,
		"pagination": gin.H{
			"total":       total,
			"page":        page,
			"limit":       limit,
			"total_pages": int(math.Ceil(float64(total) / float64(limit))),
		},
	})
}

func GetProductByID(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Id cannot be empty"})
		return
	}

	var count int
	err := config.DB.QueryRow("SELECT COUNT(*) FROM products WHERE id = ?", id).Scan(&count)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if count == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Product does not exist"})
		return
	}

	productData := models.Product{}
	err = config.DB.QueryRow(`
		SELECT id, name, price, quantity, image, sales_rate, purchase_rate 
		FROM products WHERE id = ?`, id).
		Scan(
			&productData.ID,
			&productData.Name,
			&productData.Price,
			&productData.Quantity,
			&productData.Image,
			&productData.SalesRate,
			&productData.PurchaseRate,
		)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, productData)
}

func SearchProducts(c *gin.Context) {
	// Get search query parameter
	searchQuery := c.Query("q")
	if searchQuery == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "search query parameter 'q' is required"})
		return
	}

	// Prepare the SQL query with search
	query := `
        SELECT id, name, price, quantity, image, sales_rate, purchase_rate 
        FROM products 
        WHERE name LIKE ? OR id LIKE ?
    `
	searchParam := "%" + searchQuery + "%"
	rows, err := config.DB.Query(query, searchParam, searchParam)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var products []models.Product
	for rows.Next() {
		var product models.Product
		err := rows.Scan(
			&product.ID,
			&product.Name,
			&product.Price,
			&product.Quantity,
			&product.Image,
			&product.SalesRate,
			&product.PurchaseRate,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		products = append(products, product)
	}

	if err = rows.Err(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": products,
	})
}

func CreateProduct(c *gin.Context) {
	var product models.Product

	if err := c.ShouldBindJSON(&product); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var count int
	err := config.DB.QueryRow("SELECT COUNT(*) FROM products WHERE name = ?", product.Name).Scan(&count)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if count > 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "product with this name already exists"})
		return
	}

	result, err := config.DB.Exec(`
		INSERT INTO products (name, price, quantity, image, sales_rate, purchase_rate) 
		VALUES (?, ?, ?, ?, ?, ?)`,
		product.Name,
		product.Price,
		product.Quantity,
		product.Image,
		product.SalesRate,
		product.PurchaseRate,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	id, err := result.LastInsertId()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	product.ID = uint(id)
	c.JSON(http.StatusCreated, product)
}

func UpdateProduct(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Id cannot be empty"})
		return
	}

	var product models.Product
	if err := c.ShouldBindJSON(&product); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var count int
	err := config.DB.QueryRow("SELECT COUNT(*) FROM products WHERE id = ?", id).Scan(&count)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if count == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Product does not exist"})
		return
	}

	_, err = config.DB.Exec(`
		UPDATE products 
		SET name = ?, price = ?, quantity = ?, image = ?, sales_rate = ?, purchase_rate = ?
		WHERE id = ?`,
		product.Name,
		product.Price,
		product.Quantity,
		product.Image,
		product.SalesRate,
		product.PurchaseRate,
		id,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Product updated successfully"})
}

func DeleteProduct(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Id cannot be empty"})
		return
	}

	var count int
	err := config.DB.QueryRow("SELECT COUNT(*) FROM products WHERE id = ?", id).Scan(&count)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if count == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Product does not exist"})
		return
	}

	_, err = config.DB.Exec("DELETE FROM products WHERE id = ?", id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Product deleted successfully"})
}
