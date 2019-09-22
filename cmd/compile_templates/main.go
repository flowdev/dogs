package main

import (
	"fmt"
	"log"

	"github.com/flowdev/dogs/config/bindatafs"
)

func main() {
	// Create AssetFS and name spaces
	assetFS := bindatafs.AssetFS

	// Register view paths into AssetFS
	if err := assetFS.RegisterPath("app/views/qor"); err != nil {
		log.Fatalf("Unable to register path: %v", err)
	}

	// Compile templates under registered view paths into binary
	fmt.Println("New Compile: 0")
	if err := assetFS.Compile(); err != nil {
		log.Printf("Compile err: %v", err)
	}
	fmt.Println("New Compile: 1")
}
