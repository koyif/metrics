package a

import (
	"log"
	"os"
)

func BadPanic() {
	panic("this is bad") // want "usage of panic is not allowed"
}

func BadLogFatal() {
	log.Fatal("this is bad") // want "log.Fatal should only be called in main function of main package"
}

func BadOsExit() {
	os.Exit(1) // want "os.Exit should only be called in main function of main package"
}

func NestedBad() {
	func() {
		panic("nested panic") // want "usage of panic is not allowed"
	}()
}