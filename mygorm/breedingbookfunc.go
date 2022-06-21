package mygorm

import (
	"fmt"
	"github.com/jinzhu/gorm"
)

func Init2(db *gorm.DB) error {
	if err := db.AutoMigrate(&DogTest{}, &BaseMetaFeature{}, &FeatureGroup{}, &Color{}, &SelectOneMetaFeature{}, &CheckBoxMetaFeature{},
		&StringMetaFeature{}, &TextMetaFeature{}, &IntegerMetaFeature{}, &FloatMetaFeature{}, &DateMetaFeature{}, &TimestampMetaFeature{}, &SelectManyMetaFeature{}).Error; err != nil {
		return fmt.Errorf("unable to migrate DB to current state: %v", err)
	}

	return nil
}
