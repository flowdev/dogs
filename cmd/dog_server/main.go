/*
Themen:

Naechste Schritte:
- SQLite3 upgraden: (SQLite3 >= 3.28.0) => (go-sqlite3 >= 1.10.0) => (GORM >= v1.9.11/v1.9.1?)
- Stringify() implementieren fuer ALC (fuer kuerzere Zahlen).

- Unique Index auf Namen manuell in aktueller DB erstellen.
- Schrift verbessern (unterscheidbarkeit von 1, l, I, ...)

Build with: go build -tags=bindatafs
*/
package main

import (
	"fmt"
	"html/template"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/flowdev/dogs/config/bindatafs"
	"github.com/flowdev/dogs/mygorm"
	"github.com/flowdev/dogs/myqor"
	"github.com/jinzhu/gorm"
	"github.com/mattn/go-sqlite3"
)

const generationsForTree = 6

func main() {
	workDir := filepath.Dir(os.Args[0])

	// Log to log file:
	// If the file doesn't exist, create it, or append to the file
	f, err := os.OpenFile(filepath.Join(workDir, "DogBreeding.log"), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal(err)
	}
	log.SetOutput(f)
	log.Printf("INFO: Dogs app is starting, work dir=%s", workDir)

	assetFS := bindatafs.AssetFS
	tmplContent, err := assetFS.Asset("ancestors/index.tmpl")
	if err != nil {
		log.Fatal(err)
	}
	tmplAncestors := template.Must(template.New("ancestors").Parse(string(tmplContent)))

	libVersion, _, sourceID := sqlite3.Version()
	log.Printf("INFO: sqlite3 libVersion=%s, sourceID:%s", libVersion, sourceID)
	db, err := mygorm.Init(filepath.Join(workDir, "DogBreeding.db"))
	if err != nil {
		log.Fatal(err)
	}

	// Initalize Qor Admin
	adm, err := myqor.Init(db, assetFS, workDir)
	if err != nil {
		log.Fatal(err)
	}

	ln, err := net.Listen("tcp", "127.0.0.1:8001")
	if err != nil {
		log.Fatal(err)
	}
	defer ln.Close()
	msg := fmt.Sprintf("Listening on http://%s", ln.Addr().String())
	log.Print(msg)
	fmt.Println(msg)
	mux := http.NewServeMux()
	mux.HandleFunc("/ancestors/", handleAncestors(tmplAncestors, db))
	adm.MountTo("/admin", mux)

	fmt.Println("Press 'control' + 'c' to stop the server")
	log.Fatal(http.Serve(ln, mux))
}

type tmplAncestors struct {
	Ancestors []*mygorm.Dog
	Error     error
}

func handleAncestors(tmplAncestors *template.Template, db *gorm.DB,
) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		// The "/ancestors/" pattern matches everything after this, too.
		// So we need to check that the ID (and only that) exists.
		urlParts := strings.Split(r.URL.Path, "/")
		if len(urlParts) != 3 {
			log.Printf("ERROR: Ancestor handler called for illegal path: %s", r.URL.Path)
			http.NotFound(w, r)
			return
		}
		id, err := strconv.Atoi(urlParts[2])
		if err != nil {
			log.Printf("ERROR: Dog ID '%s' isn't a valid integer: %v", urlParts[2], err)
			http.NotFound(w, r)
			return
		}
		err = tmplAncestors.Execute(w, generateAncestorTable(id, db)) // no own transaction (read only)
		if err != nil {
			log.Printf("ERROR: Unable to execute ancestor template for ID '%d': %v", id, err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

func generateAncestorTable(id int, tx *gorm.DB) tmplAncestors {
	ancestors, err := mygorm.FindAncestorsForID(tx, id, generationsForTree)
	return tmplAncestors{Ancestors: ancestors, Error: err}
}
