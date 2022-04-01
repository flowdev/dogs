# dogs
Application for dog breeding

## Building and Running

To build the application you need the following steps (all in the root directory of the project):

1. Enable vendoring: `export GO15VENDOREXPERIMENT=1`
1. Compile the templates and CSS files to Go code: `go run cmd/compile_templates/main.go`
1. Build the executable including the templates and CSS files: `go build -tags bindatafs`

Finally you can start the application (server) with: `./dogs`
