// package models

// import "gorm.io/gorm"

// type User struct {
// 	gorm.Model
// 	Name     string `json:"name" binding:"required"`
// 	Email    string `json:"email" binding:"required,email" gorm:"unique"`
// 	Password string `json:"password" binding:"required,min=6"`
// }

// func (u *User) TableName() string {
// 	return "users"
// }

package models

type User struct {
	ID       uint   `json:"id"`
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}
