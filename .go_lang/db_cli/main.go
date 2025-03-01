package main

import (
	"fmt"
)

type User struct {
	ID     uint   `gorm:"primaryKey;autoIncrement" json:"id"`
	Name   string `json:"name"`
	Status string `json:"status"`
}

func main() {
	fmt.Println("Starting DB CLI Service...")
}
