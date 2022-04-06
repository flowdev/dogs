package mygorm

import (
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/jinzhu/gorm"
)

// MateTableCount is the number of tables used for mating.
// Counting starts at 1.
const MateTableCount = 10

// UnknownHD is the display and DB value for an unknown HD.
const UnknownHD = "--"

const generationsForALC = 6
const maxCountForALC = (1 << generationsForALC) - 1

const clearMateTableSQL = "DELETE FROM mate%d;"
const countMateTableSQL = "SELECT count(*) AS count FROM mate%d;"
const listMateTableSQL = "SELECT id FROM mate%d;"
const fillMateTableSQL = `INSERT INTO mate%d (
	id, name, birth_date, alc, hd, mate_count, mother_id, father_id, remark
) SELECT
	id, name, birth_date, alc, hd, mate_count, mother_id, father_id, remark
FROM dogs WHERE star IS TRUE AND deleted_at IS NULL;`
const updateChildALCSQL = `UPDATE mate%d SET child_alc = ? WHERE ID = ?`

// Percentage can be stored in the DB and displayed nicely (with only 2 decimal places).
type Percentage float64

// String is implemented for nicer looking numbers.
func (p Percentage) String() string {
	return fmt.Sprintf("%.2f", p)
}

// Star can be stored in the DB and displayed nicely (as a star).
type Star bool

// String is implemented for nicer looking numbers.
func (s Star) String() string {
	if s {
		return "★"
	}
	return " "
}

// Scan is needed for sqlite3.
func (s *Star) Scan(src interface{}) error {
	switch v := src.(type) {
	case bool:
		*s = Star(v)
	case int64:
		*s = v != 0
	default:
		log.Printf("ERROR: Unable to scan type %T to star.", src)
		*s = false
	}
	return nil
}

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
	Name      string
	BirthDate time.Time
	ALC       Percentage
	HD        string `gorm:"size:8"`
	MateCount int
	MotherID  uint
	Mother    FemaleDog `gorm:"foreignkey:MotherID;association_autocreate:false;association_autoupdate:false"`
	FatherID  uint
	Father    MaleDog `gorm:"foreignkey:FatherID;association_autocreate:false;association_autoupdate:false"`
	ChildALC  Percentage
	Remark    string
}

// AfterDelete deletes the chick associated with the mate table after the last
// mate has been deleted.
func AfterDelete(tx *gorm.DB, tableIdx int) (bool, error) {
	count, err := countMateTable(tx, tableIdx)
	if err != nil {
		return false, err
	}
	if count <= 0 {
		if err = tx.Where("mate_table = ?", tableIdx).Delete(Chick{}).Error; err != nil {
			return false, fmt.Errorf("Unable to delete chick for mate table #%d: %v", tableIdx, err)
		}
		return true, nil
	}
	return false, nil
}

// Chick is a female dog chosen for mating (opposite of Mate).
// The technically correct term for this ('Bitch') sounds insulting for many so
// I chose 'Chick' instead.
// Deleting a record from the corresponding table should not just fill a
// deleted_at column but really delete the DB row or the same dog couldn't mate
// twice. Because of that we don't use gorm.Model.
type Chick struct {
	ID        uint `gorm:"primary_key"`
	MateALC   Percentage
	MateHD    string `gorm:"size:8"`
	MateTable int    `gorm:"unique;not null"` // we allow up to 9 tables for male partners (mate1 ... mate9)
}

// CreateChick first checks if the chick is already mating and creates it only if not.
func CreateChick(tx *gorm.DB, chick *Chick, name string) error {
	if err := CheckDoubleChick(tx, chick.ID, name); err != nil {
		return err
	}
	if err := tx.Create(chick).Error; err != nil {
		return fmt.Errorf("Unable to store chick %s: %v", name, err)
	}
	return nil
}

// CheckDoubleChick checks if the chick is already mating.
func CheckDoubleChick(tx *gorm.DB, id uint, name string) error {
	chick2 := Chick{}
	if err := tx.First(&chick2, id).Error; err != nil && !gorm.IsRecordNotFoundError(err) {
		return fmt.Errorf("Unable to find chick %s: %v", name, err)
	} else if err == nil {
		return fmt.Errorf("Chick %s is already mating in mate table %d", name, chick2.MateTable)
	}
	return nil
}

// GetChickForTable returns the current chick for the given mating table number.
func GetChickForTable(tx *gorm.DB, tableNumber string) (*Chick, error) {
	chick := Chick{}
	if err := tx.Where("mate_table = ?", tableNumber).First(&chick).Error; err != nil {
		return nil, fmt.Errorf("Unable to find chick for table %s: %v", tableNumber, err)
	}
	return &chick, nil
}

// FindFreeMateTable returns the number (between 1 and 10) of the first
// currently unused mate table.
func FindFreeMateTable(tx *gorm.DB) int {
	used := make([]bool, MateTableCount)
	chicks := []*Chick{}
	if err := tx.Find(&chicks).Error; err != nil && !gorm.IsRecordNotFoundError(err) {
		log.Printf("ERROR: Unable to list chicks: %v", err)
		return -1
	}
	if len(chicks) >= MateTableCount { // all tables are used
		return -1
	}
	for _, chick := range chicks {
		used[chick.MateTable-1] = true
	}
	for i, u := range used {
		if !u {
			return i + 1
		}
	}
	return -1 // can't happen because of early check
}

// FillMateTable fills the chosen mate table with male dogs that are candidates for mating.
func FillMateTable(tx *gorm.DB, tableIdx int, chickID uint, chickName string) error {
	sql := fmt.Sprintf(fillMateTableSQL, tableIdx) // this is safe because we know and control the source SQL
	if err := tx.Exec(sql).Error; err != nil {
		return fmt.Errorf("Unable to fill mate table %d: %v", tableIdx, err)
	}
	if count, err := countMateTable(tx, tableIdx); err != nil {
		return err
	} else if count <= 0 {
		return fmt.Errorf("No male mating partners found for %s", chickName)
	}
	if err := updateChildALCs(tx, tableIdx, chickID); err != nil {
		return err
	}
	return nil
}

// ClearMateTable deletes all mates from the chosen mate table.
func ClearMateTable(tx *gorm.DB, tableIdx int) error {
	sql := fmt.Sprintf(clearMateTableSQL, tableIdx) // this is safe because we know and control the source SQL
	if err := tx.Exec(sql).Error; err != nil {
		return fmt.Errorf("Unable to clear mate table %d: %v", tableIdx, err)
	}
	return nil
}

func countMateTable(tx *gorm.DB, tableIdx int) (int, error) {
	var count struct{ Count int }
	sql := fmt.Sprintf(countMateTableSQL, tableIdx) // this is safe because we know and control the source SQL
	if err := tx.Raw(sql).Scan(&count).Error; err != nil {
		return 0, fmt.Errorf("Unable to count mate%d table: %v", tableIdx, err)
	}
	return count.Count, nil
}

func updateChildALCs(tx *gorm.DB, tableIdx int, mumID uint) error {
	errDads := make([]uint, 0, 4096)
	updSQL := fmt.Sprintf(updateChildALCSQL, tableIdx) // this is safe because we know and control the source SQL
	d := Dog{MotherID: mumID}
	dadIDs, err := readMateIDs(tx, tableIdx)
	if err != nil {
		log.Printf("ERROR: %v", err)
		return err
	}
	for _, dadID := range dadIDs {
		var alc float64
		d.Name = fmt.Sprintf("<Potential Puppy of %d>", dadID)
		d.FatherID = dadID
		alc, err = ComputeALC(tx, &d)
		if err != nil {
			err = fmt.Errorf("Unable to compute child ALC for mate table %d, mateID %d: %v", tableIdx, dadID, err)
			log.Printf("ERROR: %v", err)
			errDads = append(errDads, dadID)
		} else if err := tx.Exec(updSQL, alc, dadID).Error; err != nil {
			err = fmt.Errorf("Unable to update child ALC in mate table %d, mateID %d: %v", tableIdx, dadID, err)
			log.Printf("ERROR: %v", err)
			errDads = append(errDads, dadID)
		}
	}
	if len(errDads) <= 1 {
		return err
	}
	err = fmt.Errorf("Updating ALCs of potential puppies failed for these potential dads: %v; last problem being: %v", errDads, err)
	log.Printf("ERROR: %v", err)
	return err
}
func readMateIDs(tx *gorm.DB, tableIdx int) ([]uint, error) {
	listSQL := fmt.Sprintf(listMateTableSQL, tableIdx) // this is safe because we know and control the source SQL
	dadIDs := make([]uint, 0, 4096)
	rows, err := tx.Raw(listSQL).Rows()
	if err != nil {
		return nil, fmt.Errorf("Unable to list mate IDs of mate table %d: %v", tableIdx, err)
	}
	defer rows.Close()
	for i := 1; rows.Next(); i++ {
		var id uint
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("Unable to read mate ID from mate table %d, row %d: %v", tableIdx, i, err)
		}
		dadIDs = append(dadIDs, id)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("Unable to read all mate IDs from mate table %d: %v", tableIdx, err)
	}
	return dadIDs, nil
}

// Litter is the result of a successful mating action.
type Litter struct {
	ID        uint `gorm:"primary_key"`
	CreatedAt time.Time
	Name      string
	ALC       Percentage
	HD        string `gorm:"size:8"`
	MotherID  uint
	Mother    FemaleDog `gorm:"foreignkey:MotherID;association_autocreate:false;association_autoupdate:false"`
	FatherID  uint
	Father    MaleDog `gorm:"foreignkey:FatherID;association_autocreate:false;association_autoupdate:false"`
	Remark    string
}

// BeforeSave is initializing the new litters HD value as soon as both parents
// are known.
func (p *Litter) BeforeSave(tx *gorm.DB) error {
	if (p.HD == "" || p.HD == UnknownHD) && p.MotherID != 0 && p.FatherID != 0 {
		m := Dog{}
		if err := tx.First(&m, p.MotherID).Error; err != nil {
			msg := fmt.Sprintf("Unable to read mother with ID '%d'.", p.MotherID)
			log.Printf("ERROR: %v", msg)
			return errors.New(msg)
		}
		f := Dog{}
		if err := tx.First(&f, p.FatherID).Error; err != nil {
			msg := fmt.Sprintf("Unable to read father with ID '%d'.", p.FatherID)
			log.Printf("ERROR: %v", msg)
			return errors.New(msg)
		}
		p.HD = CombineHD(m.HD, f.HD)
	}
	return nil
}

// Dog has got Mother and Father parents.
// This is the central data structure of the whole application.
type Dog struct {
	gorm.Model
	Name      string `gorm:"size:32;unique;not null"`
	BirthDate *time.Time
	Gender    string `gorm:"size:16"`
	Star      Star
	ALC       Percentage
	HD        string `gorm:"size:8"`
	MateCount int
	MotherID  uint
	Mother    FemaleDog `gorm:"foreignkey:MotherID;association_autocreate:false;association_autoupdate:false"`
	FatherID  uint
	Father    MaleDog `gorm:"foreignkey:FatherID;association_autocreate:false;association_autoupdate:false"`
	Remark    string
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
		d.HD = CombineHD(m.HD, f.HD)
	}
	return nil
}

// BeforeDelete checks if the dog is currently mating and returns an error if
// this is the case.
func (d *Dog) BeforeDelete(tx *gorm.DB) error {
	c := Chick{}
	if err := tx.First(&c, d.ID).Error; err != nil && !gorm.IsRecordNotFoundError(err) {
		msg := fmt.Sprintf("Unable to read chick with ID '%d': %v", d.ID, err)
		log.Printf("ERROR: %v", msg)
		return errors.New(msg)
	} else if err == nil {
		return fmt.Errorf("Dog %s is still mating in mate table %d", d.Name, c.MateTable)
	}
	var tables []int
	tx.Table("all_mates").Where("id = ?", d.ID).Pluck("mate_table", &tables)
	if len(tables) == 1 {
		return fmt.Errorf("Dog %s is still mating in mate table %d", d.Name, tables[0])
	} else if len(tables) > 1 {
		return fmt.Errorf("Dog %s is still mating in mate tables %s and %d",
			d.Name,
			joinNumbers(tables[:len(tables)-1], ", "),
			tables[len(tables)-1])
	}
	return nil
}
func joinNumbers(nums []int, sep string) string {
	b := strings.Builder{}
	for i, num := range nums {
		if i > 0 {
			b.WriteString(", ")
		}
		b.WriteString(strconv.Itoa(num))
	}
	return b.String()
}

// ComputeALC calculates the correct ALC for 6 generations.
// WARNING: All generations have to be present!
func ComputeALC(tx *gorm.DB, dog *Dog) (float64, error) {
	ancestors, err := FindAncestorsForDog(tx, dog, generationsForALC)
	if err != nil {
		return 0, err
	}
	ancestorIDSet := make(map[uint]bool)
	for i, a := range ancestors {
		if a == nil {
			msg := fmt.Sprintf("Unable to compute ALC for dog %s: Ancestor with index %d is missing", dog.Name, i)
			log.Print("ERROR: " + msg)
			return 0, errors.New(msg)
		}
		ancestorIDSet[a.ID] = true
	}
	alc := float64(len(ancestorIDSet)*100) / maxCountForALC
	return alc, nil
}

// FindAncestorsForID finds all ancestors up to the given generation.
func FindAncestorsForID(tx *gorm.DB, id int, generations int) ([]*Dog, error) {
	dog := Dog{}
	if err := tx.First(&dog, id).Error; err != nil {
		msg := fmt.Sprintf("Unable to read dog with ID '%d': %v", id, err)
		log.Printf("ERROR: %v", msg)
		return nil, errors.New(msg)
	}
	return FindAncestorsForDog(tx, &dog, generations)
}

// FindAncestorsForDog finds all ancestors up to the given generation.
func FindAncestorsForDog(tx *gorm.DB, dog *Dog, generations int) ([]*Dog, error) {
	ancestors := make([]*Dog, 0, (1<<generations)-1)
	ancestors = append(ancestors, dog)

	ancestors, err := findAncestors(tx, dog, ancestors, 1, generations)
	return ancestors, err
}

func findAncestors(tx *gorm.DB, dog *Dog, ancestors []*Dog, curGeneration, maxGeneration int,
) ([]*Dog, error) {
	m := &Dog{}
	if dog == nil || dog.MotherID == 0 {
		m = nil
	} else if err := tx.First(m, dog.MotherID).Error; gorm.IsRecordNotFoundError(err) {
		m = nil
	} else if err != nil {
		msg := fmt.Sprintf("Unable to read mother of %s with ID '%d' in generation %d: %v",
			dog.Name, dog.MotherID, curGeneration+1, err)
		log.Printf("ERROR: %v", msg)
		return nil, errors.New(msg)
	}
	ancestors = append(ancestors, m)

	if curGeneration+1 < maxGeneration {
		var err error
		ancestors, err = findAncestors(tx, m, ancestors, curGeneration+1, maxGeneration)
		if err != nil {
			return nil, err
		}
	}

	f := &Dog{}
	if dog == nil || dog.FatherID == 0 {
		f = nil
	} else if err := tx.First(f, dog.FatherID).Error; gorm.IsRecordNotFoundError(err) {
		f = nil
	} else if err != nil {
		msg := fmt.Sprintf("Unable to read father of %s with ID '%d' in generation %d: %v",
			dog.Name, dog.FatherID, curGeneration+1, err)
		log.Printf("ERROR: %v", msg)
		return nil, errors.New(msg)
	}
	ancestors = append(ancestors, f)

	curGeneration++
	if curGeneration >= maxGeneration {
		return ancestors, nil
	}

	return findAncestors(tx, f, ancestors, curGeneration, maxGeneration)
}

// CombineHD combines the two given HD values in a predictable way.
func CombineHD(hd1, hd2 string) string {
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

	if err = db.AutoMigrate(&Dog{}, &Chick{}, &Litter{}, &Mate1{}, &Mate2{}, &Mate3{},
		&Mate4{}, &Mate5{}, &Mate6{}, &Mate7{}, &Mate8{}, &Mate9{}, &Mate10{}).Error; err != nil {

		return nil, fmt.Errorf("unable to migrate DB to current state: %v", err)
	}

	if err = db.Exec(`DROP VIEW IF EXISTS female_dogs;`).Error; err != nil {
		return nil, fmt.Errorf("unable to delete view 'female_dogs': %v", err)
	}
	err = db.Exec(`CREATE VIEW female_dogs AS
	SELECT id, created_at, updated_at, deleted_at, name || ' / ' || round(alc, 2) || ' / ' || hd AS name FROM dogs
	WHERE gender = 'F' AND hd IN ('A1', 'A2', 'B1', 'B2', 'C1', 'C2');`).Error
	if err != nil {
		return nil, fmt.Errorf("unable to create view 'female_dogs': %v", err)
	}

	if err = db.Exec(`DROP VIEW IF EXISTS male_dogs;`).Error; err != nil {
		return nil, fmt.Errorf("unable to delete view 'male_dogs': %v", err)
	}
	err = db.Exec(`CREATE VIEW male_dogs AS
	SELECT id, created_at, updated_at, deleted_at, name || ' / ' || round(alc, 2) || ' / ' || hd AS name FROM dogs
	WHERE gender = 'M' AND hd IN ('A1', 'A2', 'B1', 'B2', 'C1', 'C2');`).Error
	if err != nil {
		return nil, fmt.Errorf("unable to create view 'male_dogs': %v", err)
	}

	if err = db.Exec(`DROP VIEW IF EXISTS all_mates;`).Error; err != nil {
		return nil, fmt.Errorf("unable to delete view 'all_mates': %v", err)
	}
	err = db.Exec(`CREATE VIEW all_mates AS
	SELECT id, 1 AS mate_table from mate1 UNION SELECT id, 2 AS mate_table from mate2 UNION
	SELECT id, 3 AS mate_table from mate3 UNION SELECT id, 4 AS mate_table from mate4 UNION
	SELECT id, 5 AS mate_table from mate5 UNION SELECT id, 6 AS mate_table from mate6 UNION
	SELECT id, 7 AS mate_table from mate7 UNION SELECT id, 8 AS mate_table from mate8 UNION
	SELECT id, 9 AS mate_table from mate9 UNION SELECT id, 9 AS mate_table from mate10 UNION
	SELECT id, 10 AS mate_table from mate10;`).Error
	if err != nil {
		return nil, fmt.Errorf("unable to create view 'all_mates': %v", err)
	}
	return db, nil
}

// TODO: Add more mate resources if necessary!!!

// GenericMate returns the Mate common to all mate tables.
func GenericMate(result interface{}) *Mate {
	switch mate := result.(type) {
	case *Mate1:
		return &mate.Mate
	case *Mate2:
		return &mate.Mate
	case *Mate3:
		return &mate.Mate
	case *Mate4:
		return &mate.Mate
	case *Mate5:
		return &mate.Mate
	case *Mate6:
		return &mate.Mate
	case *Mate7:
		return &mate.Mate
	case *Mate8:
		return &mate.Mate
	case *Mate9:
		return &mate.Mate
	case *Mate10:
		return &mate.Mate
	}
	panic(fmt.Sprintf("Unknown mate type: %T", result))
}

func ChickForMate(tx *gorm.DB, result interface{}) (*Chick, error) {
	var table int

	switch result.(type) {
	case *Mate1:
		table = 1
	case *Mate2:
		table = 2
	case *Mate3:
		table = 3
	case *Mate4:
		table = 4
	case *Mate5:
		table = 5
	case *Mate6:
		table = 6
	case *Mate7:
		table = 7
	case *Mate8:
		table = 8
	case *Mate9:
		table = 9
	case *Mate10:
		table = 10
	}
	if table == 0 {
		panic(fmt.Sprintf("Unknown mate type: %T", result))
	}

	return GetChickForTable(tx, strconv.Itoa(table))
}

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

// Mate10 is the 10. mate table.
type Mate10 struct {
	Mate
}
