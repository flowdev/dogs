package main

import (
	"fmt"

	"github.com/flowdev/dogs/config/bindatafs"
)

func main() {
	// Create AssetFS and name spaces
	assetFS := bindatafs.AssetFS

	// Register view paths into AssetFS
	//assetFS.RegisterPath("github.com/flowdev/dogs/app/views/qor")
	//assetFS.RegisterPath("github.com/flowdev/dogs/app/views/ancestors")
	assetFS.RegisterPath("./app/views")

	// Compile templates under registered view paths into binary
	fmt.Println("New Compile: 0")
	assetFS.Compile()
	fmt.Println("New Compile: 1")
}
