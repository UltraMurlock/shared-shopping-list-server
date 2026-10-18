package main

import (
	"fmt"
	"os"
	"shared-shopping-list-server/internal/shared/auth"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintln(os.Stderr, "usage: hashpassword <password>")
		os.Exit(1)
	}

	hash, err := auth.HashPassword(os.Args[1])
	if err != nil {
		fmt.Fprintf(os.Stderr, "hash password: %v\n", err)
		os.Exit(1)
	}

	fmt.Println(hash)
}
