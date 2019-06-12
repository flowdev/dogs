package myqor

import (
	"log"
	"sync"

	"github.com/jinzhu/gorm"
	"github.com/ole108/dogs/mygorm"
	"github.com/qor/admin"
	"github.com/qor/roles"
)

// MateData contains all data needed for initiating a single mating.
type MateData struct {
	chick *mygorm.Chick
	res   *admin.Resource
}

// mateData holds all data for mating.
var mateData = struct {
	data  []MateData
	adm   *admin.Admin
	mutex *sync.RWMutex
}{
	data:  make([]MateData, ChickCount),
	mutex: &sync.RWMutex{},
}

func getCurrentChickForTable(i int) mygorm.Chick {
	if i < 1 || i > ChickCount {
		return mygorm.Chick{}
	}
	mateData.mutex.RLock()
	defer mateData.mutex.RUnlock()
	chick := mateData.data[i].chick
	if chick == nil {
		return mygorm.Chick{}
	}
	return *chick
}

func populateChicks(db *gorm.DB) error {
	chicks := []mygorm.Chick{}
	if err := db.Find(&chicks).Error; err != nil && !gorm.IsRecordNotFoundError(err) {
		log.Printf("ERROR: Unable to load female dogs chosen for mating from DB: %v", err)
		return err
	}
	mateData.mutex.Lock()
	defer mateData.mutex.Unlock()
	for _, chick := range chicks {
		addMateChick(&chick)
	}
	return nil
}

func findFreeMateTable(chick *mygorm.Chick) int {
	for i, md := range mateData.data {
		if md.chick == nil {
			return i + 1
		}
	}
	return -1 // all tables are used
}

func addMateChick(chick *mygorm.Chick) {
	i := chick.MateTable
	if i < 1 || i > ChickCount {
		log.Printf("WARN: Chick partner table out of range: %#v", chick)
		return
	}
	mateData.data[i-1].chick = chick

	if mateData.data[i-1].res != nil {
		mateData.data[i-1].res.Name = "Mate: " + chick.Name
		mateData.data[i-1].res.Config.Invisible = false // do we have to do something to activate this???
		return
	}

	// TODO: Add more mate resources if necessary!!!
	var tbl interface{}
	switch i {
	case 1:
		tbl = &mygorm.Mate1{}
	case 2:
		tbl = &mygorm.Mate2{}
	case 3:
		tbl = &mygorm.Mate3{}
	case 4:
		tbl = &mygorm.Mate4{}
	case 5:
		tbl = &mygorm.Mate5{}
	case 6:
		tbl = &mygorm.Mate6{}
	case 7:
		tbl = &mygorm.Mate7{}
	case 8:
		tbl = &mygorm.Mate8{}
	case 9:
		tbl = &mygorm.Mate9{}
	default:
		log.Printf("ERROR: Please add missing Mate table #%d", i)
		return
	}
	mateData.data[i-1].res = mateData.adm.AddResource(tbl, &admin.Config{
		Name:       "Mate " + chick.Name,
		Invisible:  false,
		Permission: roles.Deny(roles.Update, roles.Anyone).Deny(roles.Create, roles.Anyone),
	})
}
