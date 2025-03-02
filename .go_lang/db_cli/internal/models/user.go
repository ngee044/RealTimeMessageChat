package models

type User struct {
	ID     uint   `gorm:"primaryKey;autoIncrement" json:"id"`
	Name   string `json:"name"`
	Status string `json:"status"`
}
