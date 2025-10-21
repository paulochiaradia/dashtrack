package main

import (
	"fmt"
	"os"

	"golang.org/x/crypto/bcrypt"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run tmp/generate_hash.go <password>")
		fmt.Println("Example: go run tmp/generate_hash.go MyNewPassword123")
		os.Exit(1)
	}

	password := os.Args[1]

	// Using bcrypt cost of 12 (same as the app configuration)
	hash, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		fmt.Println("Error generating hash:", err)
		os.Exit(1)
	}

	fmt.Println("\n===========================================")
	fmt.Printf("Password: %s\n", password)
	fmt.Println("===========================================")
	fmt.Printf("Hash: %s\n", string(hash))
	fmt.Println("===========================================\n")
	fmt.Println("To update in database, run:")
	fmt.Println("docker exec -it dashtrack-db-1 psql -U user -d dashtrack")
	fmt.Println("\nThen execute:")
	fmt.Printf("UPDATE users SET password = '%s' WHERE email = 'your-email@example.com';\n", string(hash))
	fmt.Println("===========================================\n")
}
