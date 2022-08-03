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

func TestBaseMetaFeatureAndFeatureGroups(t *testing.T) {
	// create colors
	c3 := &Color{Name: "blue", HexValue: "0000ff"}

	// write to DB
	if err := db.Create(c3).Error; err != nil {
		t.Fatalf("Unable to create Color 3, %v", err)
	}

	// create feature groups
	fg3 := &FeatureGroup{Name: "pigmente2", ColorID: c3.ID, ShortName: "pigments2"}
	fg4 := &FeatureGroup{Name: "teeth", ColorID: c3.ID, ShortName: "teeth"}

	// write to DB
	if err := db.Create(fg3).Error; err != nil {
		t.Fatalf("Unable to create feature group 1, %v", err)
	}

	// write to DB
	if err := db.Create(fg4).Error; err != nil {
		t.Fatalf("Unable to create feature group 2, %v", err)
	}

	// create BaseMetaFeature
	f1 := &BaseMetaFeature{Name: "pigmente2", GroupID: fg3.ID, ShortName: "pigments2", Type: TypeText}
	f2 := &BaseMetaFeature{Name: "pigmente3", GroupID: fg3.ID, ShortName: "pigments3", Type: TypeString}
	f3 := &BaseMetaFeature{Name: "pigmente4", GroupID: fg3.ID, ShortName: "pigments4", Type: TypeCheckbox}
	f4 := &BaseMetaFeature{Name: "teeth", GroupID: fg4.ID, ShortName: "teeth", Type: TypeDate}
	f5 := &BaseMetaFeature{Name: "teeth2", GroupID: fg4.ID, ShortName: "teeth2", Type: TypeFloat}

	// f1 ------------------------

	// write to DB
	if err := db.Create(f1).Error; err != nil {
		t.Fatalf("Unable to create feature 1, %v", err)
	}

	// read from DB
	f1New := &BaseMetaFeature{}
	if err := db.First(f1New, f1.ID).Error; err != nil {
		t.Fatalf("Unable to read the feature 1, %v", err)
	}

	if f1New.Type != TypeText {
		t.Errorf("f1New should have got '%s' as type but instead it is '%s': %s", TypeText, f1New.Type, spew.Sdump(f1New))
	}

	// checking the association
	if err := db.Model(f1New).Association("Group").Error; err != nil {
		t.Fatalf("Group association error: %v", err)
	}

	// population association
	db.Model(f1New).Association("Group").Find(&(f1New.Group))
	if f1New.Group.ID != fg3.ID {
		t.Errorf("f1New should have got '%s' as group but instead it is '%s': %s", fg3.Name, f1New.Group.Name, spew.Sdump(f1New))
	}

	// f2 ------------------------
	// write to DB
	if err := db.Create(f2).Error; err != nil {
		t.Fatalf("Unable to create feature 2, %v", err)
	}

	// read from DB
	f2New := &BaseMetaFeature{}
	if err := db.First(f2New, f2.ID).Error; err != nil {
		t.Fatalf("Unable to read the feature group 2, %v", err)
	}

	// checking the association
	if err := db.Model(f2New).Association("Group").Error; err != nil {
		t.Fatalf("Group association error: %v", err)
	}

	// population association
	db.Model(f2New).Association("Group").Find(&(f2New.Group))
	if f2New.Group.ID != fg3.ID {
		t.Errorf("f2New should have got '%s' as group but instead it is '%s': %s", fg3.Name, f2New.Group.Name, spew.Sdump(f2New))
	}

	// f3 ------------------------
	if err := db.Create(f3).Error; err != nil {
		t.Fatalf("Unable to create feature group 3, %v", err)
	}

	f3New := &BaseMetaFeature{}
	if err := db.First(f3New, f3.ID).Error; err != nil {
		t.Fatalf("Unable to reade the feature group 3, %v", err)
	}

	if err := db.Model(f3New).Association("Group").Error; err != nil {
		t.Fatalf("Group association error: %v", err)
	}

	db.Model(f3New).Association("Group").Find(&(f3New.Group))
	if f3New.Group.ID != fg3.ID {
		t.Errorf("f3new should have got '%s' as group but instead it is '%s': %s", fg3.Name, f3New.Group.Name, spew.Sdump(f3New))
	}

	// f4 ------------------------
	if err := db.Create(f4).Error; err != nil {
		t.Fatalf("Unable to create feature group 4, %v", err)
	}

	f4New := &BaseMetaFeature{}
	if err := db.First(f4New, f4.ID).Error; err != nil {
		t.Fatalf("Unable to reade the feature  group 4, %v", err)
	}

	if err := db.Model(f4New).Association("Group").Error; err != nil {
		t.Fatalf("Group association err: %v", err)
	}

	db.Model(f4New).Association("Group").Find(&(f4New.Group))
	if f4New.Group.ID != fg4.ID {
		t.Errorf("f4New should have got '%s' as group but instead it is '%s': %s", fg4.Name, f4New.Group.Name, spew.Sdump(f4New))
	}

	// f5 ------------------------
	if err := db.Create(f5).Error; err != nil {
		t.Fatalf("Unable to create feature group 5, %v", err)
	}

	f5New := &BaseMetaFeature{}
	if err := db.First(f5New, f5.ID).Error; err != nil {
		t.Fatalf("Unable to reade the feature group 5, %v", err)
	}

	if err := db.Model(f5New).Association("Group").Error; err != nil {
		t.Fatalf("Group association err: %v", err)
	}

	db.Model(f5New).Association("Group").Find(&(f5New.Group))
	if f5New.Group.ID != fg4.ID {
		t.Errorf("f5New should have got '%s' as group but instead it is '%s': %s", fg4.Name,
			f5New.Group.Name, spew.Sdump(f5New))
	}
}
