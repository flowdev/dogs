package myqor

import (
	"errors"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"net/url"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/flowdev/dogs/mygorm"
	"github.com/jinzhu/gorm"
	"github.com/qor/admin"
	"github.com/qor/assetfs"
	"github.com/qor/roles"
)

// MainRoute is the main route of the app.
const MainRoute = "/admin/dogs"

// MateResourcePrefix is the prefix of all QOR mate table resources.
const MateResourcePrefix = "Mate table "

const cssClassBadValue = " bad-value"

const year = time.Hour*24*365 + time.Hour*6 // 365.25 days per year
const mateMaxAge = 10 * year
const chickMaxAge = 8 * year

type dogsMateAction struct {
	ALC float64
	HD  string
}

// we have 9 mate tables: 1 ... 9
// so we don't use index 0
var mateResources [10]*admin.Resource

// Init initializes the qor admin UI by creating and configuring all resources.
func Init(db *gorm.DB, assetFS assetfs.Interface, workDir string) (*admin.Admin, error) {
	adm := admin.New(&admin.AdminConfig{DB: db, SiteName: "Dog Breeding"})
	adm.SetAssetFS(assetFS)

	/*
		// Resource for looking at the chicks
		adm.AddResource(&mygorm.Chick{}, &admin.Config{
			Invisible:  false,
			Priority:   2,
			Permission: roles.Deny(roles.Create, roles.Anyone),
		})
	*/

	// Resource for mating dialogue
	dogsMateRes := adm.NewResource(&dogsMateAction{})
	dogsMateRes.Meta(&admin.Meta{Name: "HD", Config: &admin.SelectOneConfig{Collection: []string{"A1", "A2", "B1", "B2"}}})

	// Resource for managing the dogs: MAIN RESOURCE
	dogRes := adm.AddResource(&mygorm.Dog{}, &admin.Config{
		Priority: 1,
	})
	dogRes.Meta(&admin.Meta{Name: "BirthDate", Type: "date"})
	dogRes.Meta(&admin.Meta{Name: "Gender", Config: &admin.SelectOneConfig{Collection: []string{"F", "M"}}})
	dogRes.Meta(&admin.Meta{Name: "HD", Config: &admin.SelectOneConfig{Collection: []string{mygorm.UnknownHD, "A1", "A2", "B1", "B2", "C1", "C2"}}})
	dogRes.Action(&admin.Action{
		Name: "Ancestors",
		URL: func(record interface{}, context *admin.Context) string {
			if dog, ok := record.(*mygorm.Dog); ok {
				return fmt.Sprintf("/ancestors/%d", dog.ID)
			}
			log.Printf("ERROR: Expected dog but got: %#v", record)
			return "#"
		},
		Modes: []string{"show", "edit", "menu_item"},
	})
	dogRes.Action(&admin.Action{
		Name:     "Start mating",
		Handler:  handleStartMating,
		Resource: dogsMateRes,
		Visible: func(record interface{}, context *admin.Context) bool {
			if dog, ok := record.(*mygorm.Dog); ok {
				if dog.Gender != "F" || len(dog.HD) != 2 || dog.HD == mygorm.UnknownHD {
					return false
				}
				err := mygorm.CheckDoubleChick(context.DB, dog.ID, dog.Name)
				return err == nil
			}
			return false
		},
		Modes: []string{"show", "edit", "menu_item"},
	})
	dogRes.Action(&admin.Action{
		Name:    "Compute ALC",
		Handler: handleCalculateALC,
		Modes:   []string{"batch", "show", "edit", "menu_item"},
	})

	// Resources for choosing a mate for a chick
	// TODO: Add more mate resources if necessary!!!
	mateResources[1] = adm.AddResource(&mygorm.Mate1{}, &admin.Config{
		Name: "Mate table 1", Priority: 101,
	})
	updateMateResource(mateResources[1])
	mateResources[2] = adm.AddResource(&mygorm.Mate2{}, &admin.Config{
		Name: "Mate table 2", Priority: 102,
	})
	updateMateResource(mateResources[2])
	mateResources[3] = adm.AddResource(&mygorm.Mate3{}, &admin.Config{
		Name: "Mate table 3", Priority: 103,
	})
	updateMateResource(mateResources[3])
	mateResources[4] = adm.AddResource(&mygorm.Mate4{}, &admin.Config{
		Name: "Mate table 4", Priority: 104,
	})
	updateMateResource(mateResources[4])
	mateResources[5] = adm.AddResource(&mygorm.Mate5{}, &admin.Config{
		Name: "Mate table 5", Priority: 105,
	})
	updateMateResource(mateResources[5])
	mateResources[6] = adm.AddResource(&mygorm.Mate6{}, &admin.Config{
		Name: "Mate table 6", Priority: 106,
	})
	updateMateResource(mateResources[6])
	mateResources[7] = adm.AddResource(&mygorm.Mate7{}, &admin.Config{
		Name: "Mate table 7", Priority: 107,
	})
	updateMateResource(mateResources[7])
	mateResources[8] = adm.AddResource(&mygorm.Mate8{}, &admin.Config{
		Name: "Mate table 8", Priority: 108,
	})
	updateMateResource(mateResources[8])
	mateResources[9] = adm.AddResource(&mygorm.Mate9{}, &admin.Config{
		Name: "Mate table 9", Priority: 109,
	})
	updateMateResource(mateResources[9])

	// Special Resource for a Dog that should be shown in a HTML template...
	dogTmplRes := adm.NewResource(&mygorm.Dog{}, &admin.Config{
		Name:       "dogTmplRes",
		Invisible:  false,
		Permission: roles.Allow(roles.Read, roles.Anyone),
	})
	dogTmplRes.Meta(&admin.Meta{Name: "BirthDate", Permission: roles.Allow(roles.Read, roles.Anyone), Type: "date"})
	dogTmplRes.Meta(&admin.Meta{Name: "Mother", Permission: roles.Allow(roles.Read, roles.Anyone)})
	dogTmplRes.Meta(&admin.Meta{Name: "Father", Permission: roles.Allow(roles.Read, roles.Anyone)})
	dogTmplRes.ShowAttrs("-Gender", "-Star")
	// ...and register special function to get such a Dog from the DB.
	adm.RegisterFuncMap("DogForID", getDogForID(db, dogTmplRes))
	adm.RegisterFuncMap("DogForTable", getDogForTable(db, dogTmplRes))
	adm.RegisterFuncMap("css_classes_for_value", cssClassesForValue(db))

	// Resource for Breed (the results of mating)
	breedRes := adm.AddResource(&mygorm.Breed{}, &admin.Config{
		Priority: 3,
	})
	breedRes.Meta(&admin.Meta{Name: "CreatedAt", Permission: roles.Allow(roles.Read, roles.Anyone), Type: "date"})
	breedRes.Meta(&admin.Meta{Name: "Mother", Permission: roles.Allow(roles.Read, roles.Anyone)})
	breedRes.Meta(&admin.Meta{Name: "Father", Permission: roles.Allow(roles.Read, roles.Anyone)})
	breedRes.Meta(&admin.Meta{Name: "Name", Permission: roles.Allow(roles.Read, roles.Anyone).Allow(roles.Update, roles.Anyone)})
	breedRes.Meta(&admin.Meta{Name: "Remark", Permission: roles.Allow(roles.Read, roles.Anyone).Allow(roles.Update, roles.Anyone)})

	showMateTables(db)

	removeDashboard(adm)

	return adm, nil
}

func showMateTables(tx *gorm.DB) {
	chicks := []*mygorm.Chick{}
	if err := tx.Find(&chicks).Error; err != nil && !gorm.IsRecordNotFoundError(err) {
		log.Printf("ERROR: Unable to list chicks: %v", err)
		return
	}
	for num := 1; num < 10; num++ {
		found := false
		var chick *mygorm.Chick
		for _, chick = range chicks {
			if chick.MateTable == num {
				found = true
				break
			}
		}
		if found {
			dog := mygorm.Dog{}
			if err := tx.First(&dog, chick.ID).Error; err != nil {
				log.Printf("ERROR: Unable to read dog for mate table %d (chick.ID = %d): %v", num, chick.ID, err)
			}
			setMenuIconForMateTable(num, dog.Name)
			//mateResources[chick.MateTable].Permission = roles.Allow(roles.Read, roles.Anyone).Allow(roles.Update, roles.Anyone)
		} else {
			setMenuIconForMateTable(num, "")
			//mateResources[chick.MateTable].Permission = roles.Allow(roles.Update, roles.Anyone)
		}
	}
}

func removeDashboard(adm *admin.Admin) {
	adm.GetRouter().Get("/", func(c *admin.Context) {
		http.Redirect(c.Writer, c.Request, MainRoute, http.StatusMovedPermanently)
	})
}

func handleStartMating(argument *admin.ActionArgument) error {
	// Get the user input from argument.
	dogsMateArg := argument.Argument.(*dogsMateAction)
	log.Printf("INFO: mate arg: %v", dogsMateArg)
	for _, record := range argument.FindSelectedRecords() {
		dog, ok := record.(*mygorm.Dog)
		if !ok {
			return fmt.Errorf("Expected dog but got: %T", record)
		}
		tx := argument.Context.GetDB().New() // without the `.New()` we had an old WHERE condition still set
		log.Printf("INFO: Looking to mate female dog %s.", dog.Name)
		chick := &mygorm.Chick{
			ID:        dog.ID,
			MateALC:   mygorm.Percentage(dogsMateArg.ALC),
			MateHD:    dogsMateArg.HD,
			MateTable: mygorm.FindFreeMateTable(tx),
		}
		if chick.MateTable < 0 {
			return fmt.Errorf("All mate tables are already used")
		}

		if err := mygorm.FillMateTable(tx, chick.MateTable, dog.ID, dog.Name); err != nil {
			log.Printf("ERROR: %v", err)
			return err
		}

		if err := mygorm.CreateChick(tx, chick, dog.Name); err != nil {
			log.Printf("ERROR: %v", err)
			return err
		}
		log.Printf("INFO: Chick set to: %#v", chick)

		setMenuIconForMateTable(chick.MateTable, dog.Name)
	}
	return nil
}

func handleCalculateALC(argument *admin.ActionArgument) error {
	var err error
	errDogs := make([]string, 0, 4096)
	for _, record := range argument.FindSelectedRecords() {
		dog, ok := record.(*mygorm.Dog)
		if !ok {
			err = fmt.Errorf("Expected dog but got: %T", record)
			log.Printf("ERROR: %v", err)
			return err
		}
		tx := argument.Context.GetDB().New() // without the `.New()` we had an old WHERE condition still set
		oldALC := dog.ALC
		newALC, err2 := mygorm.ComputeALC(tx, dog)
		if err2 != nil {
			err = err2
			log.Printf("ERROR: %v", err)
			errDogs = append(errDogs, dog.Name)
		} else if err2 = tx.Model(dog).Update("alc", newALC).Error; err2 != nil {
			err = fmt.Errorf("Error while updating ALC of dog %s: %v", dog.Name, err2)
			log.Printf("ERROR: %v", err)
			errDogs = append(errDogs, dog.Name)
		} else {
			log.Printf("INFO: Updated ALC for dog %s from %f to %f.", dog.Name, oldALC, newALC)
		}
	}
	if len(errDogs) <= 1 {
		return err
	}
	err = fmt.Errorf("Updating ALCs failed for these dogs: %v; last problem was: %v", errDogs, err)
	log.Printf("ERROR: %v", err)
	return err
}

func updateMateResource(mateRes *admin.Resource) {
	mateRes.Permission = roles.Allow(roles.Read, roles.Anyone).Allow(roles.Update, roles.Anyone)
	mateRes.Meta(&admin.Meta{Name: "Name", Permission: roles.Allow(roles.Read, roles.Anyone)})
	mateRes.Meta(&admin.Meta{Name: "BirthDate", Permission: roles.Allow(roles.Read, roles.Anyone), Type: "date"})
	mateRes.Meta(&admin.Meta{Name: "ALC", Permission: roles.Allow(roles.Read, roles.Anyone)})
	mateRes.Meta(&admin.Meta{Name: "HD", Permission: roles.Allow(roles.Read, roles.Anyone)})
	mateRes.Meta(&admin.Meta{Name: "MateCount", Permission: roles.Allow(roles.Read, roles.Anyone)})
	mateRes.Meta(&admin.Meta{Name: "Mother", Permission: roles.Allow(roles.Read, roles.Anyone)})
	mateRes.Meta(&admin.Meta{Name: "Father", Permission: roles.Allow(roles.Read, roles.Anyone)})
	mateRes.Meta(&admin.Meta{Name: "ChildALC", Permission: roles.Allow(roles.Read, roles.Anyone)})

	mateRes.Action(&admin.Action{
		Name:    "Remove",
		Handler: handleRemoveMates,
		Modes:   []string{"batch", "show", "edit", "menu_item"},
	})
	mateRes.Action(&admin.Action{
		Name:    "Mate!",
		Handler: handleMating,
		Modes:   []string{"show", "edit", "menu_item"},
	})
	mateRes.Action(&admin.Action{
		Name: "Ancestors",
		URL: func(record interface{}, context *admin.Context) string {
			mate := mygorm.GenericMate(record)
			return fmt.Sprintf("/ancestors/%d", mate.ID)
		},
		Modes: []string{"show", "edit", "menu_item"},
	})
}

func handleRemoveMates(argument *admin.ActionArgument) error {
	tx := argument.Context.DB
	for _, record := range argument.FindSelectedRecords() {
		if err := tx.Delete(record).Error; err != nil {
			msg := fmt.Sprintf("Unable to delete mate %#v (%T): %v", record, record, err)
			log.Print("ERROR: " + msg)
			return errors.New(msg)
		}
	}
	num, err := getMateTableNumber(argument.Context.Resource.Config.Name)
	if err != nil {
		log.Printf("ERROR: %v", err)
		return err
	}
	if del, err := mygorm.AfterDelete(tx.New(), num); err != nil {
		log.Printf("ERROR: %v", err)
		return err
	} else if del {
		setMenuIconForMateTable(num, "")
		//mateRes.Permission = roles.Allow(roles.Update, roles.Anyone)
	}
	return nil
}

func handleMating(argument *admin.ActionArgument) error {
	log.Print("DEBUG: Mating handler got called!")
	tx := argument.Context.DB
	num, err := getMateTableNumber(argument.Context.Resource.Config.Name)
	if err != nil {
		log.Printf("ERROR: %v", err)
		return err
	}
	chick := mygorm.Chick{}
	if err := tx.Where("mate_table = ?", num).First(&chick).Error; err != nil {
		msg := fmt.Sprintf("Unable to read chick for mate table %d: %v", num, err)
		log.Print("ERROR: " + msg)
		return errors.New(msg)
	}
	mum := mygorm.Dog{}
	dad := mygorm.Dog{}
	if err := tx.First(&mum, chick.ID).Error; err != nil {
		msg := fmt.Sprintf("Unable to read dog with ID %d: %v", chick.ID, err)
		log.Print("ERROR: " + msg)
		return errors.New(msg)
	}
	for _, record := range argument.FindSelectedRecords() {
		strct := reflect.ValueOf(record).Elem()
		mateValue := strct.FieldByName("Mate")
		mateIface := mateValue.Interface()
		if mate, ok := mateIface.(mygorm.Mate); ok {
			p := mygorm.Breed{
				Name:     mum.Name + " + " + mate.Name,
				ALC:      (mum.ALC + mate.ALC) / 2,
				HD:       mygorm.CombineHD(mum.HD, mate.HD),
				Remark:   "Mum: " + mum.Remark + "\nDad: " + mate.Remark,
				MotherID: mum.ID,
				FatherID: mate.ID,
			}
			if err := tx.Create(&p).Error; err != nil {
				msg := fmt.Sprintf("Unable to store breed %s: %v", p.Name, err)
				log.Print("ERROR: " + msg)
				return errors.New(msg)
			}
			if err := tx.First(&dad, mate.ID).Error; err != nil {
				msg := fmt.Sprintf("Unable to read dog with ID %d: %v", mate.ID, err)
				log.Print("ERROR: " + msg)
				return errors.New(msg)
			}
		} else {
			msg := fmt.Sprintf("Unable to work with non-mate %#v (%T)", record, mateIface)
			log.Print("ERROR: " + msg)
			return errors.New(msg)
		}
	}
	mum.MateCount++
	if err := tx.Save(&mum).Error; err != nil {
		msg := fmt.Sprintf("Unable to save mum %s: %v", mum.Name, err)
		log.Print("ERROR: " + msg)
		return errors.New(msg)
	}
	dad.MateCount++
	if err := tx.Save(&dad).Error; err != nil {
		msg := fmt.Sprintf("Unable to save dad %s: %v", dad.Name, err)
		log.Print("ERROR: " + msg)
		return errors.New(msg)
	}
	if err := mygorm.ClearMateTable(tx, num); err != nil {
		log.Printf("ERROR: %v", err)
		return err
	}
	if _, err := mygorm.AfterDelete(tx, num); err != nil {
		log.Printf("ERROR: %v", err)
		return err
	}
	setMenuIconForMateTable(num, "")
	//mateRes.Permission = roles.Allow(roles.Update, roles.Anyone)
	return nil
}

func setMenuIconForMateTable(mateTable int, name string) {
	res := mateResources[mateTable]
	menus := res.GetAdmin().GetMenus()
	for _, m := range menus {
		if m.Priority == 100+mateTable {
			m.IconName = name
			return
		}
	}
}

// TemplateDog is a dog suitable for showing in a HTML template.
type TemplateDog struct {
	Dogs []*mygorm.Dog
	Res  *admin.Resource
}

// getDogForID returns a function that gets a dog from URL values that contain the ID of the dog.
func getDogForID(db *gorm.DB, res *admin.Resource) func(url.Values) TemplateDog {
	return func(urlVals url.Values) TemplateDog {
		ids := urlVals[":dog_id"]
		if len(ids) <= 0 {
			log.Printf("ERROR: No dog_id found in URL values: %#v", urlVals)
			return TemplateDog{}
		}
		id, err := strconv.Atoi(ids[0])
		if err != nil {
			log.Printf("ERROR: dog_id '%s' is not a number: %v", ids[0], err)
			return TemplateDog{}
		}
		dog := mygorm.Dog{}
		if err := db.First(&dog, id).Error; err != nil {
			log.Printf("ERROR: Unable to get dog with ID '%d' from the DB: %v", id, err)
			return TemplateDog{}
		}
		return TemplateDog{Dogs: []*mygorm.Dog{&dog}, Res: res}
	}
}

func getDogForTable(db *gorm.DB, res *admin.Resource) func(string) TemplateDog {
	return func(urlPath string) TemplateDog {
		log.Printf("DEBUG: getDogForTable urlPath = %#v", urlPath)
		// extract mate table index from path
		num, err := getMateTableNumber(urlPath)
		if err != nil {
			log.Printf("ERROR: %v", err)
			return TemplateDog{}
		}
		// get chick for that table
		chick := mygorm.Chick{}
		if err := db.Where("mate_table = ?", num).First(&chick).Error; gorm.IsRecordNotFoundError(err) {
			log.Printf("DEBUG: Unable to find chick for mate table %d", num)
			return TemplateDog{}
		} else if err != nil {
			log.Printf("ERROR: Unable to get chick for mate table %d from the DB: %v", num, err)
			return TemplateDog{}
		}
		// get dog for chick
		dog := mygorm.Dog{}
		if err := db.First(&dog, chick.ID).Error; err != nil {
			log.Printf("ERROR: Unable to get dog with ID '%d' from the DB: %v", chick.ID, err)
			return TemplateDog{}
		}
		return TemplateDog{Dogs: []*mygorm.Dog{&dog}, Res: res}
	}
}

func getMateTableNumber(name string) (int, error) {
	if len(name) <= 0 {
		return 0, fmt.Errorf("No mate table found in: %q", name)
	}
	mateTable := name[len(name)-1:]
	// convert it to int
	num, err := strconv.Atoi(mateTable)
	if err != nil {
		return 0, fmt.Errorf("mateTable '%s' (of %q) is not a number: %v", mateTable, name, err)
	}
	return num, nil
}

func cssClassesForValue(db *gorm.DB) func(value, result interface{}, fieldName, resName string) template.HTML {
	return func(value, result interface{}, fieldName, resName string) template.HTML {
		if strings.HasPrefix(resName, MateResourcePrefix) {
			chick, err := mygorm.GetChickForTable(db, resName[len(MateResourcePrefix):])
			if err != nil {
				log.Printf("ERROR: %v", err)
				return ""
			}
			mate := mygorm.GenericMate(result)
			switch fieldName {
			case "ChildALC":
				if mate.ChildALC < chick.MateALC {
					return cssClassBadValue
				}
			case "HD":
				if mate.HD > chick.MateHD {
					return cssClassBadValue
				}
			case "BirthDate":
				now := time.Now()
				if mate.BirthDate.Add(mateMaxAge).Before(now) {
					return cssClassBadValue
				}
			}
		} else if resName == "dogTmplRes" && fieldName == "BirthDate" {
			chick, ok := result.(*mygorm.Dog)
			if !ok {
				log.Printf("ERROR: Expected result type *mygorm.Dog but got: %T", result)
				return ""
			}
			now := time.Now()
			if chick.BirthDate.Add(chickMaxAge).Before(now) {
				return cssClassBadValue
			}
		} else {
			//log.Printf("DEBUG: Unknown resource: %s, field: %s, result type: %T or value: %#v", resName, fieldName, result, value)
		}
		return ""
	}
}
