package myqor

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strconv"

	"github.com/flowdev/dogs/mygorm"
	"github.com/jinzhu/gorm"
	"github.com/qor/admin"
	"github.com/qor/roles"
)

// MainRoute is the main route of the app.
const MainRoute = "/admin/dogs"

// ChickCount is the maximum number of parallel matings.
const ChickCount = 9

const fillMateTableSQL = `INSERT INTO mate%ds (
	id, created_at, updated_at, name, birth_date, alc, hd,
	mate_count, mother_id, father_id
) SELECT
	id, created_at, updated_at, name, birth_date, alc, hd,
	mate_count, mother_id, father_id
FROM dogs WHERE star IS TRUE AND alc <= ? AND hd <= ?;`

type dogsMateAction struct {
	ALC float64
	HD  string
}

// Init initializes the qor admin UI by creating and configuring all resources.
func Init(db *gorm.DB) (*admin.Admin, error) {
	adm := admin.New(&admin.AdminConfig{DB: db, SiteName: "Dog Breeding"})

	dogsMateRes := adm.NewResource(&dogsMateAction{})

	chickRes := adm.AddResource(&mygorm.Chick{}, &admin.Config{
		Invisible:  false,
		Permission: roles.Deny(roles.Create, roles.Anyone),
	})
	chickRes.Meta(&admin.Meta{Name: "Name", Permission: roles.Allow(roles.Read, roles.Anyone)})
	chickRes.Meta(&admin.Meta{Name: "BirthDate", Permission: roles.Allow(roles.Read, roles.Anyone), Type: "date"})
	chickRes.Meta(&admin.Meta{Name: "ALC", Permission: roles.Allow(roles.Read, roles.Anyone)})
	chickRes.Meta(&admin.Meta{Name: "HD", Permission: roles.Allow(roles.Read, roles.Anyone)})
	chickRes.Meta(&admin.Meta{Name: "MateCount", Permission: roles.Allow(roles.Read, roles.Anyone)})
	chickRes.Meta(&admin.Meta{Name: "Mother", Permission: roles.Allow(roles.Read, roles.Anyone)})
	chickRes.Meta(&admin.Meta{Name: "Father", Permission: roles.Allow(roles.Read, roles.Anyone)})
	chickRes.Meta(&admin.Meta{Name: "MateHD", Config: &admin.SelectOneConfig{Collection: []string{"A1", "A2", "B1", "B2"}}})
	//chickRes.EditAttrs("-MateTable")
	chickRes.ShowAttrs("-MateTable")
	chickRes.IndexAttrs("-MateTable")
	chickRes.Action(&admin.Action{
		Name: "Search Mate Partners",
		Handler: func(actionArgument *admin.ActionArgument) error {
			for _, record := range actionArgument.FindSelectedRecords() {
				chick := record.(*mygorm.Chick)
				log.Printf("INFO: Chick searches partner with ALC %v and HD %s", chick.MateALC, chick.MateHD)
				//context.DB.Model(chick).Update("MateCount", chick.MateCount+1)
				// TODO: Search and insert mate partners into mate table!!!
				// TODO: Update resource MateX to be visible with right name.
			}
			return nil
		},
		Modes: []string{"show", "edit", "menu_item"},
	})

	dogRes := adm.AddResource(&mygorm.Dog{})
	dogRes.Meta(&admin.Meta{Name: "BirthDate", Type: "date"})
	dogRes.Meta(&admin.Meta{Name: "Gender", Config: &admin.SelectOneConfig{Collection: []string{"F", "M"}}})
	dogRes.Meta(&admin.Meta{Name: "HD", Config: &admin.SelectOneConfig{Collection: []string{mygorm.UnknownHD, "A1", "A2", "B1", "B2", "C1", "C2"}}})
	dogRes.Action(&admin.Action{
		Name: "Ancestors",
		URL: func(record interface{}, context *admin.Context) string {
			if dog, ok := record.(*mygorm.Dog); ok {
				return fmt.Sprintf("/ancestors/%d", dog.ID)
			}
			return "#"
		},
		Modes: []string{"show", "edit", "menu_item"},
	})
	dogRes.Action(&admin.Action{
		Name: "Mate",
		Handler: func(argument *admin.ActionArgument) error {
			// Get the user input from argument.
			dogsMateArg := argument.Argument.(*dogsMateAction)
			log.Printf("INFO: mate arg: %v", dogsMateArg)
			for _, record := range argument.FindSelectedRecords() {
				if dog, ok := record.(*mygorm.Dog); ok {
					log.Printf("INFO: Looking to mate female dog %d.", dog.ID)
					chick := &mygorm.Chick{
						Name:      dog.Name,
						BirthDate: dog.BirthDate,
						ALC:       dog.ALC,
						HD:        dog.HD,
						MateCount: dog.MateCount,
						MotherID:  dog.MotherID,
						Mother:    dog.Mother,
						FatherID:  dog.FatherID,
						Father:    dog.Father,
					}
					chick.ID = dog.ID
					chick.CreatedAt = dog.CreatedAt
					chick.UpdatedAt = dog.UpdatedAt
					chick.MateTable = findFreeMateTable(chick)
					if chick.MateTable < 0 {
						return fmt.Errorf("all mate tables are already used")
					}
					addMateChick(chick)
					if err := argument.Context.GetDB().Save(chick).Error; err != nil {
						return fmt.Errorf("unable to save chick '%s': %v", chick.Name, err)
					}
					log.Printf("INFO: Chick set to: %v", chick)
				} else {
					return fmt.Errorf("expected dog but got: %T", record)
				}
			}
			return nil
		},
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

	err := populateChicks(db)
	if err != nil {
		return nil, err
	}

	// we don't want the dashboard:
	adm.GetRouter().Get("/", func(c *admin.Context) {
		http.Redirect(c.Writer, c.Request, MainRoute, http.StatusMovedPermanently)
	})

	dogTmplRes := adm.NewResource(&mygorm.Dog{}, &admin.Config{
		Invisible:  false,
		Permission: roles.Allow(roles.Read, roles.Anyone),
	})
	dogTmplRes.Meta(&admin.Meta{Name: "BirthDate", Permission: roles.Allow(roles.Read, roles.Anyone), Type: "date"})
	dogTmplRes.Meta(&admin.Meta{Name: "Mother", Permission: roles.Allow(roles.Read, roles.Anyone)})
	dogTmplRes.Meta(&admin.Meta{Name: "Father", Permission: roles.Allow(roles.Read, roles.Anyone)})
	adm.RegisterFuncMap("Dog", getDogForID(db, dogTmplRes))

	return adm, nil
}

type TemplateDog struct {
	Dog *mygorm.Dog
	Res *admin.Resource
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
		return TemplateDog{Dog: &dog, Res: res}
	}
}
