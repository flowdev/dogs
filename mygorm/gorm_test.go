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
	m := &Dog{Name: "Blacky", HD: "A1", Gender: "F"}
	f := &Dog{Name: "Rex", HD: "A2", Gender: "M"}

	// write to DB
	if err := db.Create(m).Error; err != nil {
		t.Fatalf("Unable to create the Mother, %v", err)
	}

	if err := db.Create(f).Error; err != nil {
		t.Fatalf("Unable to create the Father, %v", err)
	}

	p1.MotherID = m.ID
	p1.FatherID = f.ID

	if err := db.Create(p1).Error; err != nil {
		t.Fatalf("Unable to create Puppy 1, %v", err)
	}

	p2.MotherID = m.ID
	p2.FatherID = f.ID

	if err := db.Create(p2).Error; err != nil {
		t.Fatalf("Unable to create Puppy 2, %v", err)
	}

	// read from DB
	p1New := &Dog{}
	if err := db.First(p1New, p1.ID).Error; err != nil {
		t.Fatalf("Unable to read puppy1 with ID %d: %v", p1.ID, err)
	}

	if err := db.Model(p1New).Association("Mother").Error; err != nil {
		t.Fatalf("Mother association error: %v", err)
	}
	db.Model(p1New).Association("Mother").Find(&(p1New.Mother))
	if p1New.Mother.ID != m.ID {
		t.Errorf("p1New should have got '%s' as mother but instead it is '%s': %s", m.Name, p1New.Mother.Name, spew.Sdump(p1New))
	}
	if err := db.Model(p1New).Association("Father").Error; err != nil {
		t.Fatalf("Father association error: %v", err)
	}
	db.Model(p1New).Association("Father").Find(&(p1New.Father))
	if p1New.Father.ID != f.ID {
		t.Errorf("p1New should have got '%s' as father but instead it is '%s': %s", f.Name, p1New.Father.Name, spew.Sdump(p1New))
	}

	p2New := &Dog{}
	if err := db.First(p2New, p2.ID).Error; err != nil {
		t.Fatalf("Unable to read puppy2 with ID %d: %v", p2.ID, err)
	}

	if err := db.Model(p2New).Association("Mother").Error; err != nil {
		t.Fatalf("Mother association error: %v", err)
	}
	db.Model(p2New).Association("Mother").Find(&(p2New.Mother))
	if p2New.Mother.ID != m.ID {
		t.Errorf("p2New should have got '%s' as mother but instead it is '%s': %s", m.Name, p2New.Mother.Name, spew.Sdump(p2New))
	}
	if err := db.Model(p2New).Association("Father").Error; err != nil {
		t.Fatalf("Father association error: %v", err)
	}
	db.Model(p2New).Association("Father").Find(&(p2New.Father))
	if p2New.Father.ID != f.ID {
		t.Errorf("p2New should have got '%s' as father but instead it is '%s': %s", f.Name, p2New.Father.Name, spew.Sdump(p2New))
	}
}

func TestColorAndFeatureGroups(t *testing.T) {
	// create colors
	c1 := &Color{Name: "red", HexValue: "ff0000"}
	c2 := &Color{Name: "green", HexValue: "00ff00"}

	// write to DB
	if err := db.Create(c1).Error; err != nil {
		t.Fatalf("Unable to create Color 1, %v", err)
	}

	if err := db.Create(c2).Error; err != nil {
		t.Fatalf("Unable to create Color 2, %v", err)
	}

	// create feature groups
	fg1 := &FeatureGroup{Name: "pigmente", ColorID: c1.ID, ShortName: "pigments"}
	fg2 := &FeatureGroup{Name: "Teeth", ColorID: c2.ID, ShortName: "Teeth"}

	// write to DB
	if err := db.Create(fg1).Error; err != nil {
		t.Fatalf("Unable to create feature group 1, %v", err)
	}

	if err := db.Create(fg2).Error; err != nil {
		t.Fatalf("Unable to create feature group 2, %v", err)
	}

	// read from DB
	fg1New := &FeatureGroup{}
	if err := db.First(fg1New, fg1.ID).Error; err != nil {
		t.Fatalf("Unable to read the feature group 1 , %v", err)
	}

	// checking the association
	if err := db.Model(fg1New).Association("Color").Error; err != nil {
		t.Fatalf("Color association error: %v", err)
	}

	// population association
	db.Model(fg1New).Association("Color").Find(&(fg1New.Color))
	if fg1New.Color.ID != c1.ID {
		t.Errorf("p1New should have got '%s' as color but instead it is '%s': %s", c1.Name, fg1New.Color.Name, spew.Sdump(fg1New))
	}

	fg2New := &FeatureGroup{}
	// read from DB
	if err := db.First(fg2New, fg2.ID).Error; err != nil {
		t.Fatalf("Unable to read the feature group 2 , %v", err)
	}

	// checking the association
	if err := db.Model(fg2New).Association("Color").Error; err != nil {
		t.Fatalf("Color association error: %v", err)
	}

	// population association
	db.Model(fg2New).Association("Color").Find(&(fg2New.Color))
	if fg2New.Color.ID != c2.ID {
		t.Errorf("p2New should have got '%s' as color but instead it is '%s': %s", c2.Name, fg2New.Color.Name, spew.Sdump(fg2New))
	}

}
