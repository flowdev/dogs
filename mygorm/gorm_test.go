package mygorm

import (
	"log"
	"os"
	"testing"

	"github.com/davecgh/go-spew/spew"
	"github.com/jinzhu/gorm"
	_ "github.com/mattn/go-sqlite3"
)

var db *gorm.DB

func init() {
	var err error
	dbFname := "test.db"

	if err = os.RemoveAll(dbFname); err != nil {
		log.Fatalf("Unable to remove database file '%s': %v", dbFname, err)
	}
	if db, err = Init(dbFname); err != nil {
		log.Fatalf("Unable to initialize database '%s': %v", dbFname, err)
	}
}

func TestDogsAndParents(t *testing.T) {
	// create dogs
	p1 := &Dog{Name: "Lilly"}
	p2 := &Dog{Name: "Hushy"}
	m := &Dog{Name: "Blacky"}
	f := &Dog{Name: "Rex"}

	// write to DB
	db.Create(m)
	db.Create(f)
	p1.MotherID = m.ID
	p1.FatherID = f.ID
	db.Create(p1)
	p2.MotherID = m.ID
	p2.FatherID = f.ID
	db.Create(p2)

	// read from DB
	p1New := &Dog{}
	db.First(p1New, p1.ID)
	if db.Model(p1New).Association("Mother").Error != nil {
		t.Fatalf("Mother association error: %v", db.Model(p1New).Association("Mother").Error)
	}
	db.Model(p1New).Association("Mother").Find(&(p1New.Mother))
	if p1New.Mother.ID != m.ID {
		t.Errorf("p1New should have got '%s' as mother but instead it is '%s': %s", m.Name, p1New.Mother.Name, spew.Sdump(p1New))
	}
	if db.Model(p1New).Association("Father").Error != nil {
		t.Fatalf("Father association error: %v", db.Model(p1New).Association("Father").Error)
	}
	db.Model(p1New).Association("Father").Find(&(p1New.Father))
	if p1New.Father.ID != f.ID {
		t.Errorf("p1New should have got '%s' as father but instead it is '%s': %s", f.Name, p1New.Father.Name, spew.Sdump(p1New))
	}

	p2New := &Dog{}
	db.First(p2New, p2.ID)
	if db.Model(p2New).Association("Mother").Error != nil {
		t.Fatalf("Mother association error: %v", db.Model(p2New).Association("Mother").Error)
	}
	db.Model(p2New).Association("Mother").Find(&(p2New.Mother))
	if p2New.Mother.ID != m.ID {
		t.Errorf("p2New should have got '%s' as mother but instead it is '%s': %s", m.Name, p2New.Mother.Name, spew.Sdump(p2New))
	}
	if db.Model(p2New).Association("Father").Error != nil {
		t.Fatalf("Father association error: %v", db.Model(p2New).Association("Father").Error)
	}
	db.Model(p2New).Association("Father").Find(&(p2New.Father))
	if p2New.Father.ID != f.ID {
		t.Errorf("p2New should have got '%s' as father but instead it is '%s': %s", f.Name, p2New.Father.Name, spew.Sdump(p2New))
	}
}
