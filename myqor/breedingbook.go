package myqor

import (
	"github.com/flowdev/dogs/mygorm"
	"github.com/jinzhu/gorm"
	"github.com/qor/admin"
)

/*
Init initializes the qor admin UI by creating and configuring all resources.
*/

func Init2(db *gorm.DB, adm *admin.Admin) error {

	// Resource for managing the dogs: MAIN RESOURCE
	dogRes := adm.AddResource(&mygorm.DogTest{}, &admin.Config{
		Priority: 19,
	})
	_ = dogRes

	featureGroupRes := adm.AddResource(&mygorm.FeatureGroup{}, &admin.Config{
		Priority: 21,
	})
	_ = featureGroupRes

	colorRes := adm.AddResource(&mygorm.Color{}, &admin.Config{
		Priority: 22,
	})
	_ = colorRes

	selectOneMetaFeatureRes := adm.AddResource(&mygorm.SelectOneMetaFeature{}, &admin.Config{
		Priority: 23,
	})
	_ = selectOneMetaFeatureRes

	checkBoxMetaFeatureRes := adm.AddResource(&mygorm.CheckBoxMetaFeature{}, &admin.Config{
		Priority: 24,
	})
	_ = checkBoxMetaFeatureRes

	stringMetaFeatureRes := adm.AddResource(&mygorm.StringMetaFeature{}, &admin.Config{
		Priority: 25,
	})
	_ = stringMetaFeatureRes

	textMetaFeatureRes := adm.AddResource(&mygorm.TextMetaFeature{}, &admin.Config{
		Priority: 26,
	})
	_ = textMetaFeatureRes

	integerMetaFeatureRes := adm.AddResource(&mygorm.IntegerMetaFeature{}, &admin.Config{
		Priority: 27,
	})
	_ = integerMetaFeatureRes

	floatMetaFeatureRes := adm.AddResource(&mygorm.FloatMetaFeature{}, &admin.Config{
		Priority: 28,
	})
	_ = floatMetaFeatureRes

	dateMetaFeatureRes := adm.AddResource(&mygorm.DateMetaFeature{}, &admin.Config{
		Priority: 29,
	})
	_ = dateMetaFeatureRes

	timestampMetaFeatureRes := adm.AddResource(&mygorm.TimestampMetaFeature{}, &admin.Config{
		Priority: 30,
	})
	_ = timestampMetaFeatureRes

	/*selectManyMetaFeatureRes := adm.AddResource(&mygorm.SelectManyMetaFeature{}, &admin.Config{
		Priority: 30,
	})
	_ = selectManyMetaFeatureRes*/

	/*attributes := []interface{}{"Name", "SelectOne1", "SelectOne2", "CheckBox1", "CheckBox2", "String1", "String1Quality", "String2", "String2Quality", "Text1", "Text1Quality", "Text2", "Text2Quality", "Integer1", "Integer2", "Float1", "Float2", "Date1", "Date2", "Timestamp1", "Timestamp2", "SelectMany1", "SelectMany2"}

	// find attributes dynamically

	// show given attributes
	dogRes.IndexAttrs("ID", "Name")

	// Set attributes will be shown in the new page
	dogRes.NewAttrs(attributes...)

	// Set attributes will be shown for the edit page, similar to new page
	dogRes.EditAttrs(attributes...)

	// generate Meta information
	dogRes.Meta(&admin.Meta{Name: "BirthDate", Type: "date"})
	dogRes.Meta(&admin.Meta{Name: "Gender", Config: &admin.SelectOneConfig{Collection: []string{"F", "M"}}})
	dogRes.Meta(&admin.Meta{Name: "HD", Config: &admin.SelectOneConfig{Collection: []string{mygorm.UnknownHD, "A1", "A2", "B1", "B2", "C1", "C2"}}})*/

	return nil
}
