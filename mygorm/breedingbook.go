package mygorm

import (
	"github.com/jinzhu/gorm"
	"time"
)

type TypeEnum string

const (
	TypeUnknown    = TypeEnum("")
	TypeString     = TypeEnum("string")
	TypeText       = TypeEnum("text")
	TypeCheckbox   = TypeEnum("checkbox")
	TypeInt        = TypeEnum("integer")
	TypeFloat      = TypeEnum("floating point")
	TypeDate       = TypeEnum("date")
	TypeDateTime   = TypeEnum("timestamp")
	TypeSelectOne  = TypeEnum("select one")
	TypeSelectMany = TypeEnum("select many") // []string
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
	BaseID     uint  `gorm:"not null"`
	NeutralMin int64 `gorm:"not null"`
	NeutralMax int64 `gorm:"not null"`
	BadMin     int64 `gorm:"not null"`
	BadMax     int64 `gorm:"not null"`
	PerfectMin int64 `gorm:"not null"`
	PerfectMax int64 `gorm:"not null"`
}

type IntegerFeature struct {
	FeatureID uint  `gorm:"not null"`
	DogID     uint  `gorm:"not null"`
	Value     int64 `gorm:"not null"`
}

type FloatMetaFeature struct {
	BaseID     uint    `gorm:"not null"`
	NeutralMin float64 `gorm:"not null"`
	NeutralMax float64 `gorm:"not null"`
	BadMin     float64 `gorm:"not null"`
	BadMax     float64 `gorm:"not null"`
	PerfectMin float64 `gorm:"not null"`
	PerfectMax float64 `gorm:"not null"`
}

type FloatFeature struct {
	FeatureID uint    `gorm:"not null"`
	DogID     uint    `gorm:"not null"`
	Value     float64 `gorm:"not null"`
}

type DateMetaFeature struct {
	BaseID uint `gorm:"not null"`
	// still decide
}

type DateFeature struct {
	FeatureID uint      `gorm:"not null"`
	DogID     uint      `gorm:"not null"`
	Value     time.Time `gorm:"not null"`
}

type TimestampMetaFeature struct {
	BaseID uint `gorm:"not null"`
	// still decide
}

type TimestampFeature struct {
	FeatureID uint      `gorm:"not null"`
	DogID     uint      `gorm:"not null"`
	Value     time.Time `gorm:"not null"`
}
