package main

import (
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

func main() {
	password := "Driver@123"
	hash, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	fmt.Println("Hash for 'Driver@123':")
	fmt.Println(string(hash))
}
