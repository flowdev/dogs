/*
Themen:
- Geburtsdatum! Maximales Alter: Weibchen: 8, Maennchen: 10 Jahre
- IK-Berechnung mit interpolation!
- AVK-Berechnung mit AVK der Ahnen (wenn bekannt)!
  Fehlende Generationen ausgleichen????!
- HD-Wert-Eingabe: Schlechtester zugelassener Wert des Rueden
- 9 Paarungen parallel anbahnen reicht???

Naechste Schritte:
- SQLite3 upgraden: >= 3.28.0 (Neues Release erst hier (Fix schon in master): https://github.com/mattn/go-sqlite3/releases dann GORM)
- Nach Auswahl der Ruedin und fuellen der Partner-Werte (AVG + HD),
  potentielle Partner in mateX fuellen.
- Was soll nach Auswahl des Rueden passieren? Ungeborenen Hund anlegen!
  Die Anzahl der Paarungen dokumentieren!
*/
package main

import (
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	"github.com/flowdev/dogs/mygorm"
	"github.com/flowdev/dogs/myqor"
	"github.com/mattn/go-sqlite3"
)

var tmplAncestors *template.Template

func init() {
	tmplContent, err := ioutil.ReadFile("./app/views/ancestors/index.tmpl")
	if err != nil {
		log.Fatal(err)
	}
	tmplAncestors = template.Must(template.New("ancestors").Parse(string(tmplContent)))
}

func main() {
	libVersion, _, sourceID := sqlite3.Version()
	log.Printf("INFO: sqlite3 libVersion=%s, sourceID:%s", libVersion, sourceID)
	db, err := mygorm.Init("dogs.db")
	if err != nil {
		log.Fatal(err)
	}

	// Initalize Qor Admin
	adm, err := myqor.Init(db)
	if err != nil {
		log.Fatal(err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/ancestors/", handleAncestors)
	//mux.HandleFunc("/new-chick/", handleNewChick(db))
	adm.MountTo("/admin", mux)

	port := 9000
	log.Printf("Listening on port: %d", port)
	defer http.ListenAndServe(fmt.Sprintf(":%d", port), mux)
}

func handleAncestors(w http.ResponseWriter, r *http.Request) {
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

func generateAncestorTable(id string) template.HTML {
	// TODO: find ancestors from DB
	return template.HTML(id)
}
