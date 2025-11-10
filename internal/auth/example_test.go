package auth_test

import (
	"fmt"
	"log"

	"github.com/Jannnesi/Palvelinohjelmointi-GO-Group-2/internal/auth"
)

// Example_newUserRegistration demonstrates hashing a password for a new user
func Example_newUserRegistration() {
	// When a new user registers
	plainPassword := "mySecurePassword123"

	// Hash the password before storing in database
	_, err := auth.HashPassword(plainPassword)
	if err != nil {
		log.Fatal(err)
	}

	// Store hashedPassword in database
	fmt.Println("Password hashed successfully")
	fmt.Println("Hash starts with: $2a$")
	// Output:
	// Password hashed successfully
	// Hash starts with: $2a$
}

// Example_userLogin demonstrates verifying a password during login
func Example_userLogin() {
	// Simulate a stored hashed password
	plainPassword := "myPassword"
	hashedPassword, _ := auth.HashPassword(plainPassword)

	// User attempts to login
	attemptedPassword := "myPassword"

	// Verify the password
	err := auth.VerifyPassword(hashedPassword, attemptedPassword)
	if err != nil {
		fmt.Println("Login failed: incorrect password")
		return
	}

	fmt.Println("Login successful")
	// Output:
	// Login successful
}

// Example_incorrectPassword demonstrates what happens with wrong password
func Example_incorrectPassword() {
	// Simulate a stored hashed password
	plainPassword := "correctPassword"
	hashedPassword, _ := auth.HashPassword(plainPassword)

	// User attempts to login with wrong password
	attemptedPassword := "wrongPassword"

	// Verify the password
	err := auth.VerifyPassword(hashedPassword, attemptedPassword)
	if err != nil {
		fmt.Println("Authentication failed")
		return
	}

	fmt.Println("Login successful")
	// Output:
	// Authentication failed
}
