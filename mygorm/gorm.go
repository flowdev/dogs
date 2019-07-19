package mygorm

import (
	"fmt"
	"log"
	"time"

	"github.com/jinzhu/gorm"
)

// UnknownHD is the display and DB value for an unknown HD.
const UnknownHD = "--"

// FemaleDog is a view of female dogs (rows in the dogs table with gender set
// to 'F').
// A FemaleDog belongs to a Dog (it's child) as Mother.
type FemaleDog struct {
	gorm.Model
	Name string
}

// MaleDog is a view of male dogs (rows in the dogs table with gender set to
// 'M').
// A MaleDog belongs to a Dog (it's child) as Father.
type MaleDog struct {
	gorm.Model
	Name string
}

// Mate is a male dog chosen as a candidate for mating.
// Mate is like Dog but doesn't need Star or Gender because it is always true
// and male.
// Deleting a record from the corresponding table should not just fill a
// deleted_at column but really delete the DB row or the same dog couldn't mate
// twice. Because of that we don't use gorm.Model.
type Mate struct {
	ID        uint `gorm:"primary_key"`
	CreatedAt time.Time
	UpdatedAt time.Time
	Name      string
	BirthDate *time.Time
	ALC       float64
	HD        string `gorm:"size:8"`
	MateCount int
	MotherID  uint
	Mother    FemaleDog `gorm:"foreignkey:MotherID;association_autocreate:false;association_autoupdate:false"`
	FatherID  uint
	Father    MaleDog `gorm:"foreignkey:FatherID;association_autocreate:false;association_autoupdate:false"`
}

// Chick is a female dog chosen for mating (opposite of Mate).
// The technically correct term for this ('Bitch') sounds insulting for many so
// I chose 'Chick' instead.
// Deleting a record from the corresponding table should not just fill a
// deleted_at column but really delete the DB row or the same dog couldn't mate
// twice. Because of that we don't use gorm.Model.
type Chick struct {
	ID        uint `gorm:"primary_key"`
	CreatedAt time.Time
	UpdatedAt time.Time
	Name      string `gorm:"size:32"`
	BirthDate *time.Time
	ALC       float64
	HD        string `gorm:"size:8"`
	MateCount int
	MotherID  uint
	Mother    FemaleDog `gorm:"foreignkey:MotherID;association_autocreate:false;association_autoupdate:false"`
	FatherID  uint
	Father    MaleDog `gorm:"foreignkey:FatherID;association_autocreate:false;association_autoupdate:false"`
	MateALC   float64
	MateHD    string `gorm:"size:8"`
	MateTable int    `gorm:"unique;not null"` // we allow up to 9 tables for male partners (mate1 ... mate9)
}

// FindFreeMateTable returns the number (between 1 and 9) of the first
// currently unused mate table.
func FindFreeMateTable() int {
	for i, md := range mateData.data {
		if md.chick == nil {
			return i + 1
		}
	}
	return -1 // all tables are used
}

// Dog has got Mother and Father parents.
// This is the central data structure of the whole application.
type Dog struct {
	gorm.Model
	Name      string `gorm:"size:32"`
	BirthDate *time.Time
	Gender    string `gorm:"size:16"`
	Star      bool
	ALC       float64
	HD        string `gorm:"size:8"`
	MateCount int
	MotherID  uint
	Mother    FemaleDog `gorm:"foreignkey:MotherID;association_autocreate:false;association_autoupdate:false"`
	FatherID  uint
	Father    MaleDog `gorm:"foreignkey:FatherID;association_autocreate:false;association_autoupdate:false"`
}

// BeforeSave is initializing the new dogs HD value as soon as both parents are
// known.
func (d *Dog) BeforeSave(tx *gorm.DB) error {
	if (d.HD == "" || d.HD == UnknownHD) && d.MotherID != 0 && d.FatherID != 0 {
		m := Dog{}
		if err := tx.First(&m, d.MotherID).Error; err != nil {
			log.Printf("ERROR: Unable to read mother with ID '%d'.", d.MotherID)
			return err
		}
		f := Dog{}
		if err := tx.First(&f, d.FatherID).Error; err != nil {
			log.Printf("ERROR: Unable to read father with ID '%d'.", d.FatherID)
			return err
		}
		d.HD = combineALC(m.HD, f.HD)
	}
	return nil
}
func combineALC(hd1, hd2 string) string {
	if hd1 <= hd2 {
		return hd1 + " " + hd2
	}
	return hd2 + " " + hd1
}

// Init is initializing the DB.
func Init(dbFname string) (*gorm.DB, error) {
	log.Printf("Opening database '%s'...", dbFname)
	db, err := gorm.Open("sqlite3", dbFname)
	if err != nil {
		return nil, fmt.Errorf("the database '%s' could not be opened", dbFname)
	}

	if err = db.AutoMigrate(&Dog{}, &Mate{}, &Chick{}, &Mate1{}, &Mate2{},
		&Mate3{}, &Mate4{}, &Mate5{}, &Mate6{}, &Mate7{}, &Mate8{}, &Mate9{}).Error; err != nil {

		return nil, fmt.Errorf("unable to migrate DB to current state: %v", err)
	}

	if err = db.Exec(`DROP VIEW IF EXISTS female_dogs;`).Error; err != nil {
		return nil, fmt.Errorf("unable to delete view 'female_dogs': %v", err)
	}
	err = db.Exec(`CREATE VIEW female_dogs AS
	SELECT id, created_at, updated_at, deleted_at, name || ' / ' || CAST(alc AS text) || ' / ' || hd AS name FROM dogs
	WHERE gender = 'F' AND hd IN ('A1', 'A2', 'B1', 'B2', 'C1', 'C2');`).Error
	if err != nil {
		return nil, fmt.Errorf("unable to create view 'female_dogs': %v", err)
	}

	if err = db.Exec(`DROP VIEW IF EXISTS male_dogs;`).Error; err != nil {
		return nil, fmt.Errorf("unable to delete view 'male_dogs': %v", err)
	}
	err = db.Exec(`CREATE VIEW male_dogs AS
	SELECT id, created_at, updated_at, deleted_at, name || ' / ' || CAST(alc AS text) || ' / ' || hd AS name FROM dogs
	WHERE gender = 'M' AND hd IN ('A1', 'A2', 'B1', 'B2', 'C1', 'C2');`).Error
	if err != nil {
		return nil, fmt.Errorf("unable to create view 'male_dogs': %v", err)
	}
	return db, nil
}

// TODO: Add more mate resources if necessary!!!

// Mate1 is the 1. mate table.
type Mate1 struct {
	Mate
}

// Mate2 is the 2. mate table.
type Mate2 struct {
	Mate
}

// Mate3 is the 3. mate table.
type Mate3 struct {
	Mate
}

// Mate4 is the 4. mate table.
type Mate4 struct {
	Mate
}

// Mate5 is the 5. mate table.
type Mate5 struct {
	Mate
}

// Mate6 is the 6. mate table.
type Mate6 struct {
	Mate
}

// Mate7 is the 7. mate table.
type Mate7 struct {
	Mate
}

// Mate8 is the 8. mate table.
type Mate8 struct {
	Mate
}

// Mate9 is the 9. mate table.
type Mate9 struct {
	Mate
}
