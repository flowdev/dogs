/*
Themen:

Naechste Schritte:
- Schrift verbessern per font-family in: app/views/qor/layout.tmpl
- SQLite3 upgraden: (SQLite3 >= 3.28.0) => (go-sqlite3 >= 1.10.0) => (GORM >= v1.9.11/v1.9.1?)

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
	db.SetLogger(log.New(f, "\n", 0))

	// Initalize QOR Admin
	adm, err := myqor.Init(db, assetFS, workDir)
	if err != nil {
		log.Fatal(err)
	}

	ln, err := net.Listen("tcp", "127.0.0.1:8001")
	if err != nil {
		log.Fatal(err)
	}
	defer ln.Close()
	msg := fmt.Sprintf("Listening on http://%s/admin", ln.Addr().String())
	log.Print(msg)
	fmt.Println(msg)
	mux := http.NewServeMux()
	mux.HandleFunc("/ancestors/", handleAncestors(tmplAncestors, db))
	adm.MountTo("/admin", mux)

	fmt.Println("Press 'control' + 'c' to stop the server")
	log.Fatal(http.Serve(ln, mux))
}

type tmplAncestors struct {
	Ancestors  []*mygorm.Dog
	Quantities []int //Number of times the ancestors appear in the tree
	Error      error
}

func handleAncestors(tmplAncestors *template.Template, db *gorm.DB,
) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var id2 int

		// The "/ancestors/" pattern matches everything after this, too.
		// So we need to check that the ID (and only that) exists.
		urlParts := strings.Split(r.URL.Path, "/")
		n := len(urlParts)
		if n < 3 || n > 4 {
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
		if n > 3 {
			id2, err = strconv.Atoi(urlParts[3])
			if err != nil {
				log.Printf("ERROR: Dog ID (2) '%s' isn't a valid integer: %v", urlParts[3], err)
				http.NotFound(w, r)
				return
			}
		} else {
			id2 = -1
		}
		err = tmplAncestors.Execute(w, generateAncestorTable(id, id2, db)) // no own transaction (read only)
		if err != nil {
			log.Printf("ERROR: Unable to execute ancestor template for ID '%d': %v", id, err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

func generateAncestorTable(id, id2 int, tx *gorm.DB) tmplAncestors {
	var ancestors []*mygorm.Dog
	var err error

	if id2 < 0 {
		ancestors, err = mygorm.FindAncestorsForID(tx, id, generationsForTree)
	} else {
		d := &mygorm.Dog{
			MotherID: uint(id),
			FatherID: uint(id2),
			Name:     "Puppy",
		}
		alc, err2 := mygorm.ComputeALC(tx, d)
		if err2 != nil {
			log.Printf("ERROR: Unable to calculate the ALC for Puppy of %d and %d: %v", id, id2, err2)
		} else {
			d.ALC = mygorm.Percentage(alc)
		}
		ancestors, err = mygorm.FindAncestorsForDog(tx, d, generationsForTree)
	}

	quantities := calculateQuantities(ancestors)

	return tmplAncestors{
		Ancestors:  ancestors,
		Quantities: quantities,
		Error:      err,
	}
}

func calculateQuantities(ancestors []*mygorm.Dog) []int {
	tmpDogCount := make(map[uint]int, len(ancestors))
	for _, dog := range ancestors {
		tmpDogCount[dog.ID]++
	}

	quantities := make([]int, len(ancestors))
	for i, dog := range ancestors {
		quantities[i] = tmpDogCount[dog.ID]
	}
	return quantities
}
