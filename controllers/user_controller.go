package controllers

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/umesh/ginapi/config"
	"github.com/umesh/ginapi/models"
)

func GetUsers(c *gin.Context) {
	rows, err := config.DB.Query(`
        SELECT id, name, email, created_at, updated_at 
        FROM users
    `)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var users []models.User
	for rows.Next() {
		var user models.User
		err := rows.Scan(
			&user.ID,
			&user.Name,
			&user.Email,
			&user.CreatedAt,
			&user.UpdatedAt,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		users = append(users, user)
	}

	if err = rows.Err(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, users)
}

func GetUsersByID(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Id cannot be empty"})
		return
	}

	var user models.User
	err := config.DB.QueryRow(`
        SELECT id, name, email, created_at, updated_at 
        FROM users 
        WHERE id = ?`, id,
	).Scan(
		&user.ID,
		&user.Name,
		&user.Email,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, user)
}

func CreateUser(c *gin.Context) {
	var user models.User

	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	fmt.Println("---------------")
	fmt.Println(user.Email)
	fmt.Println(user.Name)
	fmt.Println(user.Password)
	fmt.Println("---------------")
	if user.Name == "" || user.Email == "" || user.Password == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "name, email and password are required"})
		return
	}

	var count int
	err := config.DB.QueryRow(`
        SELECT COUNT(*) FROM users WHERE email = ?`,
		user.Email,
	).Scan(&count)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if count > 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "email already exists"})
		return
	}

	hashedPassword := hashPassword(user.Password)

	result, err := config.DB.Exec(`
        INSERT INTO users (name, email, password) 
        VALUES (?, ?, ?)`,
		user.Name, user.Email, hashedPassword,
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

	newUser := models.User{}
	err = config.DB.QueryRow(`
        SELECT id, name, email, created_at, updated_at 
        FROM users 
        WHERE id = ?`, id,
	).Scan(
		&newUser.ID,
		&newUser.Name,
		&newUser.Email,
		&newUser.CreatedAt,
		&newUser.UpdatedAt,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, newUser)
}

func UpdateUser(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user ID is required"})
		return
	}

	tokenUserID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	tokenUserIDStr := string(fmt.Sprintf("%v", tokenUserID))
	if id != tokenUserIDStr {
		c.JSON(http.StatusForbidden, gin.H{"error": "you can only update your own profile"})
		return
	}

	var user models.User
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var currentUser models.User
	err := config.DB.QueryRow(`
        SELECT email FROM users WHERE id = ?`, id,
	).Scan(&currentUser.Email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	if user.Email != currentUser.Email {
		var count int
		err := config.DB.QueryRow(`
            SELECT COUNT(*) FROM users 
            WHERE email = ? AND id != ?`,
			user.Email, id,
		).Scan(&count)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		if count > 0 {
			c.JSON(http.StatusConflict, gin.H{"error": "email already in use"})
			return
		}
	}

	var hashedPassword string
	if user.Password != "" {
		hashedPassword = hashPassword(user.Password)
	} else {
		err := config.DB.QueryRow(`
            SELECT password FROM users WHERE id = ?`, id,
		).Scan(&hashedPassword)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}

	_, err = config.DB.Exec(`
        UPDATE users 
        SET name = ?, email = ?, password = ? 
        WHERE id = ?`,
		user.Name, user.Email, hashedPassword, id,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	updatedUser := models.User{}
	err = config.DB.QueryRow(`
        SELECT id, name, email, created_at, updated_at 
        FROM users 
        WHERE id = ?`, id,
	).Scan(
		&updatedUser.ID,
		&updatedUser.Name,
		&updatedUser.Email,
		&updatedUser.CreatedAt,
		&updatedUser.UpdatedAt,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, updatedUser)
}
