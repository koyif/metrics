package main

import (
	"log"
	"os"
)

// Good: log.Fatal and os.Exit are allowed in main function of main package
func main() {
	if false {
		log.Fatal("allowed here")
	}
	if false {
		os.Exit(1)
	}
}

// Bad: log.Fatal outside main function
func helper() {
	log.Fatal("not allowed") // want "log.Fatal should only be called in main function of main package"
}

// Bad: os.Exit outside main function
func anotherHelper() {
	os.Exit(1) // want "os.Exit should only be called in main function of main package"
}

// Bad: panic is never allowed
func panicHelper() {
	panic("never allowed") // want "usage of panic is not allowed"
}