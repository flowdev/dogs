package mygorm

import (
	"fmt"
	"github.com/jinzhu/gorm"
)

/*
Init initializes the qor admin UI by creating and configuring all resources.
*/

func Init2(db *gorm.DB) error {
	// Set up the database
	if err := db.AutoMigrate(&DogTest{}, &FeatureGroup{}, &Color{}, &SelectOneMetaFeature{}, &CheckBoxMetaFeature{},
		&StringMetaFeature{}, &TextMetaFeature{}, &IntegerMetaFeature{}, &FloatMetaFeature{}, &DateMetaFeature{}, &TimestampMetaFeature{}).Error; err != nil {
		return fmt.Errorf("unable to migrate DB to current state: %v", err)
	}

	// Initalize

	// Create resources from GORM-backend model

	return nil
}
