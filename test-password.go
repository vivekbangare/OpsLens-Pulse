package main

import (
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

func main() {
	password := "admin" // 🔥 Change this to whatever you want

	// Cost 12 is production safe (same as your DB hash)
	hash, err := bcrypt.GenerateFromPassword(
		[]byte(password),
		12,
	)

	if err != nil {
		panic(err)
	}

	fmt.Println("Password:", password)
	fmt.Println("Bcrypt Hash:")
	fmt.Println(string(hash))
}
