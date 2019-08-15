package myqor

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"reflect"
	"strconv"

	"github.com/flowdev/dogs/mygorm"
	"github.com/jinzhu/gorm"
	"github.com/qor/admin"
	"github.com/qor/qor"
	"github.com/qor/roles"
)

// MainRoute is the main route of the app.
const MainRoute = "/admin/dogs"

type dogsMateAction struct {
	ALC float64
	HD  string
}

// we have 9 mate tables: 1 ... 9
// so we don't use index 0
var mateResources [10]*admin.Resource

// Init initializes the qor admin UI by creating and configuring all resources.
func Init(db *gorm.DB) (*admin.Admin, error) {
	adm := admin.New(&admin.AdminConfig{DB: db, SiteName: "Dog Breeding"})

	// Resource for looking at the chicks
	adm.AddResource(&mygorm.Chick{}, &admin.Config{
		Invisible:  false,
		Priority:   2,
		Permission: roles.Deny(roles.Create, roles.Anyone),
	})

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
				return dog.Gender == "F" && len(dog.HD) == 2 &&
					dog.HD != mygorm.UnknownHD
			}
			return false
		},
		Modes: []string{"show", "edit", "menu_item"},
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

	// Resource for Puppies (the results of mating)
	adm.AddResource(&mygorm.Puppy{}, &admin.Config{
		Priority:   3,
		Permission: roles.Deny(roles.Create, roles.Anyone),
	})

	//adjustMateMenus()

	//showMateTables(db)

	removeDashboard(adm)

	return adm, nil
}

func adjustMateMenus() {
	for _, mr := range mateResources {
		if mr != nil {
			mr.Permission = roles.Allow(roles.Update, roles.Anyone)
		}
	}
}

func showMateTables(tx *gorm.DB) {
	chicks := []*mygorm.Chick{}
	if err := tx.Find(&chicks).Error; err != nil && !gorm.IsRecordNotFoundError(err) {
		log.Printf("ERROR: Unable to list chicks: %v", err)
		return
	}
	for _, chick := range chicks {
		dog := mygorm.Dog{}
		tx.First(&dog, chick.ID)
		//setMenuNameForMateTable(chick.MateTable, "Mating "+dog.Name)
		mateResources[chick.MateTable].Permission = roles.Allow(roles.Read, roles.Anyone).Allow(roles.Update, roles.Anyone)
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
			MateALC:   dogsMateArg.ALC,
			MateHD:    dogsMateArg.HD,
			MateTable: mygorm.FindFreeMateTable(tx),
		}
		if chick.MateTable < 0 {
			return fmt.Errorf("All mate tables are already used")
		}

		if err := mygorm.CreateChick(tx, chick, dog.Name); err != nil {
			log.Printf("ERROR: %v", err)
			return err
		}
		log.Printf("INFO: Chick set to: %#v", chick)

		if err := mygorm.FillMateTable(tx, chick.MateTable, chick.MateALC, chick.MateHD, dog.Name); err != nil {
			log.Printf("ERROR: %v", err)
			return err
		}

		//mateResources[chick.MateTable].Permission = roles.Allow(roles.Read, roles.Anyone).Allow(roles.Update, roles.Anyone)
		//setMenuNameForMateTable(chick.MateTable, "Mating "+dog.Name)
	}
	return nil
}

func updateMateResource(mateRes *admin.Resource) {
	mateRes.Permission = roles.Allow(roles.Read, roles.Anyone).Allow(roles.Update, roles.Anyone)
	mateRes.Meta(&admin.Meta{Name: "BirthDate", Permission: roles.Allow(roles.Read, roles.Anyone), Type: "date"})
	mateRes.Meta(&admin.Meta{Name: "Mother", Permission: roles.Allow(roles.Read, roles.Anyone)})
	mateRes.Meta(&admin.Meta{Name: "Father", Permission: roles.Allow(roles.Read, roles.Anyone)})
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
	c := mygorm.Chick{}
	if err := tx.Where("mate_table = ?", num).First(&c).Error; err != nil {
		msg := fmt.Sprintf("Unable to read chick for mate table %d: %v", num, err)
		log.Print("ERROR: " + msg)
		return errors.New(msg)
	}
	chick := mygorm.Dog{}
	if err := tx.First(&chick, c.ID).Error; err != nil {
		msg := fmt.Sprintf("Unable to read chick for mate table %d: %v", num, err)
		log.Print("ERROR: " + msg)
		return errors.New(msg)
	}
	for _, record := range argument.FindSelectedRecords() {
		strct := reflect.ValueOf(record).Elem()
		mateValue := strct.FieldByName("Mate")
		mateIface := mateValue.Interface()
		if mate, ok := mateIface.(mygorm.Mate); ok {
			p := mygorm.Puppy{
				Name:     chick.Name + " + " + mate.Name,
				ALC:      (chick.ALC + mate.ALC) / 2,
				HD:       mygorm.CombineHD(chick.HD, mate.HD),
				MotherID: chick.ID,
				FatherID: mate.ID,
			}
			if err := tx.Create(&p).Error; err != nil {
				msg := fmt.Sprintf("Unable to store puppy %s: %v", p.Name, err)
				log.Print("ERROR: " + msg)
				return errors.New(msg)
			}
		} else {
			msg := fmt.Sprintf("Unable to work with non-mate %#v (%T)", record, mateIface)
			log.Print("ERROR: " + msg)
			return errors.New(msg)
		}
		if err := mygorm.ClearMateTable(tx, num); err != nil {
			log.Printf("ERROR: %v", err)
			return err
		}
	}
	if _, err := mygorm.AfterDelete(tx.New(), num); err != nil {
		log.Printf("ERROR: %v", err)
		return err
	}
	//mateRes.Permission = roles.Allow(roles.Update, roles.Anyone)
	return nil
}

func setMenuNameForMateTable(mateTable int, name string) {
	res := mateResources[mateTable]
	menus := res.GetAdmin().GetMenus()
	for _, m := range menus {
		if m.Priority == 100+mateTable {
			m.Name = name
			res.Name = name
			return
		}
	}
}

type menuPermissioner struct {
	menu *admin.Menu
}

// HasPermission implements the qor/admin Permissioner interface
func (mp menuPermissioner) HasPermission(mode roles.PermissionMode, ctx *qor.Context) bool {
	return string(mode) == "read" && mp.menu.Name != ""
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
		idx, err := getMateTableNumber(urlPath)
		if err != nil {
			log.Printf("ERROR: %v", err)
			return TemplateDog{}
		}
		// get chick for that table
		chick := mygorm.Chick{}
		db.Where("mate_table = ?", idx).First(&chick)
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
