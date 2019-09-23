/*
Themen:
- Geburtsdatum! Maximales Alter: Weibchen: 8, Maennchen: 10 Jahre
- AVK-Berechnung mit AVK der Ahnen (wenn bekannt)!
  Fehlende Generationen ausgleichen????!
- HD-Wert-Eingabe: Schlechtester zugelassener Wert des Rueden
- 9 Paarungen parallel anbahnen reicht???

Naechste Schritte:
- SQLite3 upgraden: >= 3.28.0 (Neues Release erst hier (Fix schon in master): https://github.com/mattn/go-sqlite3/releases dann GORM)
*/
package main

import (
	"html/template"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/flowdev/dogs/config/bindatafs"
	"github.com/flowdev/dogs/mygorm"
	"github.com/flowdev/dogs/myqor"
	"github.com/mattn/go-sqlite3"
	"github.com/zserge/webview"
)


func main() {
	workDir := filepath.Join(filepath.Dir(os.Args[0]), "..", "Data")
	if err := os.Chdir(workDir); err != nil {
		log.Fatal(err)
	}

	// Log to log file:
	// If the file doesn't exist, create it, or append to the file
	f, err := os.OpenFile("DogBreeding.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		log.Fatal(err)
	}
	log.SetOutput(f)
	log.Printf("INFO: Dogs app is starting, work dir=%s", workDir)

	// Register view paths into AssetFS
	assetFS := bindatafs.AssetFS
	tmplContent, err := assetFS.Asset("ancestors/index.tmpl")
	if err != nil {
		log.Fatal(err)
	}
	tmplAncestors := template.Must(template.New("ancestors").Parse(string(tmplContent)))

	libVersion, _, sourceID := sqlite3.Version()
	log.Printf("INFO: sqlite3 libVersion=%s, sourceID:%s", libVersion, sourceID)
	db, err := mygorm.Init("DogBreeding.db")
	if err != nil {
		log.Fatal(err)
	}

	// Initalize Qor Admin
	adm, err := myqor.Init(db, assetFS)
	if err != nil {
		log.Fatal(err)
	}

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		log.Fatal(err)
	}
	defer ln.Close()
	go func() {
		log.Printf("Listening on http://%s", ln.Addr().String())
		mux := http.NewServeMux()
		mux.HandleFunc("/ancestors/", handleAncestors(tmplAncestors))
		adm.MountTo("/admin", mux)

		log.Fatal(http.Serve(ln, mux))
	}()
	webview.Open("Dog Breeding", "http://"+ln.Addr().String()+"/admin/dogs", 1200, 800, true)
}

func handleAncestors(tmplAncestors *template.Template,
) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		// The "/ancestors/" pattern matches everything after this, too.
		// So we need to check that the ID (and only that) exists.
		urlParts := strings.Split(r.URL.Path, "/")
		if len(urlParts) != 3 {
			http.NotFound(w, r)
			return
		}
		err := tmplAncestors.Execute(w, generateAncestorTable(urlParts[2]))
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

func generateAncestorTable(id string) template.HTML {
	// TODO: find ancestors from DB
	return template.HTML(id)
}
