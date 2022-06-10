package mygorm

import (
	"github.com/jinzhu/gorm"
)

type TypeEnum string

const (
	TypeUnknown    = TypeEnum("")
	TypeString     = TypeEnum("string")         // done
	TypeText       = TypeEnum("text")           // done
	TypeCheckbox   = TypeEnum("checkbox")       // done
	TypeInt        = TypeEnum("integer")        // int64
	TypeFloat      = TypeEnum("floating point") //float64
	TypeDate       = TypeEnum("date")           //time.Time
	TypeDateTime   = TypeEnum("timestamp")      //time.Time
	TypeSelectOne  = TypeEnum("select one")     // string done
	TypeSelectMany = TypeEnum("select many")    // []string
)

var AllTypeEnums = []TypeEnum{
	TypeUnknown,
	TypeString,
	TypeText,
	TypeCheckbox,
	TypeInt,
	TypeFloat,
	TypeDate,
	TypeDateTime,
	TypeSelectOne,
	TypeSelectMany,
}

type QualityEnum string

const (
	QualityNeutral = QualityEnum("neutral")
	QualityBad     = QualityEnum("bad")
	QualityPerfect = QualityEnum("perfect")
)

var AllQualityEnums = []QualityEnum{
	QualityNeutral,
	QualityBad,
	QualityPerfect,
}

type BaseMetaFeature struct {
	gorm.Model
	Name      string       `gorm:"unique; not null"`
	ShortName string       `gorm:"unique; not null"`
	Type      TypeEnum     `gorm:"not null"`
	GroupID   uint         `gorm:"not null"`
	Group     FeatureGroup `gorm:"foreignkey:GroupID;association_autocreate:false;association_autoupdate:false"`
}

type FeatureGroup struct {
	gorm.Model
	Name      string `gorm:"unique; not null"`
	ShortName string `gorm:"unique; not null"`
	ColorID   uint   `gorm:"not null"`
	Color     Color  `gorm:"foreignkey:ColorID;association_autocreate:false;association_autoupdate:false"`
}

type Color struct {
	gorm.Model
	Name     string `gorm:"unique; not null"`
	HexValue string `gorm:"unique; not null"`
}

type SelectOneMetaFeature struct {
	BaseID  uint        `gorm:"not null"`
	Value   string      `gorm:"not null"`
	Quality QualityEnum `gorm:"not null"`
}

type SelectOneFeature struct {
	FeatureID uint   `gorm:"not null"`
	DogID     uint   `gorm:"not null"`
	Value     string `gorm:"not null"`
}

type CheckBoxMetaFeature struct {
	BaseID  uint        `gorm:"not null"`
	Value   bool        `gorm:"not null"`
	Quality QualityEnum `gorm:"not null"`
}

type CheckBoxFeature struct {
	FeatureID uint `gorm:"not null"`
	DogID     uint `gorm:"not null"`
	Value     bool `gorm:"not null"`
}

type StringMetaFeature struct {
	BaseID    uint `gorm:"not null"`
	MinLength uint `gorm:"not null"`
	MaxLength uint `gorm:"not null"`
}

type StringFeature struct {
	FeatureID uint        `gorm:"not null"`
	DogID     uint        `gorm:"not null"`
	Value     string      `gorm:"not null"`
	Quality   QualityEnum `gorm:"not null"`
}

type TextMetaFeature struct {
	BaseID    uint `gorm:"not null"`
	MinLength uint `gorm:"not null"`
	MaxLength uint `gorm:"not null"`
}

type TextFeature struct {
	FeatureID uint        `gorm:"not null"`
	DogID     uint        `gorm:"not null"`
	Value     string      `gorm:"not null"`
	Quality   QualityEnum `gorm:"not null"`
}

type IntegerMetaFeature struct {
	BaseID      uint        `gorm:"not null"`
	Value       int64       `gorm:"not null"`
	SmallerThan bool        `gorm:"not null"`
	Quality     QualityEnum `gorm:"not null"`
}

type IntegerFeature struct {
	FeatureID uint  `gorm:"not null"`
	DogID     uint  `gorm:"not null"`
	Value     int64 `gorm:"not null"`
}
