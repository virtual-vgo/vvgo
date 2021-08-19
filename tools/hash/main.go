package main

import (
	"fmt"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/crypto/ssh/terminal"
	"log"
	"os"
)

func main() {
	fmt.Print("enter password: ")
	password1, err := terminal.ReadPassword(int(os.Stdin.Fd()))
	fmt.Println()
	if err != nil {
		log.Fatalf("terminal.ReadPassword() failed: %v", err)
	}

	hash, err := bcrypt.GenerateFromPassword(password1, bcrypt.DefaultCost)
	if err != nil {
		log.Fatalf("bcrypt.GenerateFromPassword() failed: %v", err)
	}

	fmt.Print("enter same password: ")
	password2, err := terminal.ReadPassword(int(os.Stdin.Fd()))
	fmt.Println()
	if err != nil {
		log.Fatalf("terminal.ReadPassword() failed: %v", err)
	}

	if err := bcrypt.CompareHashAndPassword(hash, password2); err != nil {
		log.Fatalf("bcrypt.CompareHashAndPassword() failed: %v", err)
	}

	fmt.Println("hash:", string(hash))
	os.Exit(0)
}
