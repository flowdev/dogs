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

type DogTest struct {
	gorm.Model
	Name           string      `gorm:"unique; not null"`
	SelectOne1     string      `gorm:"not null"`
	SelectOne2     string      `gorm:"not null"`
	SelectOne3     string      `gorm:"not null"`
	CheckBox1      bool        `gorm:"not null"`
	CheckBox2      bool        `gorm:"not null"`
	CheckBox3      bool        `gorm:"not null"`
	String1        string      `gorm:"not null"`
	String2        string      `gorm:"not null"`
	String3        string      `gorm:"not null"`
	String1Quality QualityEnum `gorm:"not null"`
	String2Quality QualityEnum `gorm:"not null"`
	String3Quality QualityEnum `gorm:"not null"`
	Text1          string      `gorm:"not null"`
	Text2          string      `gorm:"not null"`
	Text3          string      `gorm:"not null"`
	Text1Quality   QualityEnum `gorm:"not null"`
	Text2Quality   QualityEnum `gorm:"not null"`
	Text3Quality   QualityEnum `gorm:"not null"`
	Integer1       int64       `gorm:"not null"`
	Integer2       int64       `gorm:"not null"`
	Integer3       int64       `gorm:"not null"`
	Float1         float64     `gorm:"not null"`
	Float2         float64     `gorm:"not null"`
	Float3         float64     `gorm:"not null"`
	Date1          time.Time   `gorm:"not null"`
	Date2          time.Time   `gorm:"not null"`
	Date3          time.Time   `gorm:"not null"`
	Timestamp1     time.Time   `gorm:"not null"`
	Timestamp2     time.Time   `gorm:"not null"`
	Timestamp3     time.Time   `gorm:"not null"`
	SelectMany1    []string    `gorm:"not null"`
	SelectMany2    []string    `gorm:"not null"`
	SelectMany3    []string    `gorm:"not null"`
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
	BaseID    uint        `gorm:"not null"`
	ColumnNum int         `gorm:"not null"`
	Order     int         `gorm:"not null"`
	Value     string      `gorm:"not null"`
	Quality   QualityEnum `gorm:"not null"`
}

type CheckBoxMetaFeature struct {
	BaseID           uint        `gorm:"not null"`
	ColumnNum        int         `gorm:"not null"`
	QualityChecked   QualityEnum `gorm:"not null"`
	QualityUnchecked QualityEnum `gorm:"not null"`
}

type StringMetaFeature struct {
	BaseID    uint `gorm:"not null"`
	ColumnNum int  `gorm:"not null"`
	MinLength uint `gorm:"not null"`
	MaxLength uint `gorm:"not null"`
}

type TextMetaFeature struct {
	BaseID    uint `gorm:"not null"`
	ColumnNum int  `gorm:"not null"`
	MinLength uint `gorm:"not null"`
	MaxLength uint `gorm:"not null"`
}

type IntegerMetaFeature struct {
	BaseID     uint  `gorm:"not null"`
	ColumnNum  int   `gorm:"not null"`
	NeutralMin int64 `gorm:"not null"`
	NeutralMax int64 `gorm:"not null"`
	BadMin     int64 `gorm:"not null"`
	BadMax     int64 `gorm:"not null"`
	PerfectMin int64 `gorm:"not null"`
	PerfectMax int64 `gorm:"not null"`
}

type FloatMetaFeature struct {
	BaseID     uint    `gorm:"not null"`
	ColumnNum  int     `gorm:"not null"`
	NeutralMin float64 `gorm:"not null"`
	NeutralMax float64 `gorm:"not null"`
	BadMin     float64 `gorm:"not null"`
	BadMax     float64 `gorm:"not null"`
	PerfectMin float64 `gorm:"not null"`
	PerfectMax float64 `gorm:"not null"`
}

type DateMetaFeature struct {
	BaseID    uint `gorm:"not null"`
	ColumnNum int  `gorm:"not null"`
	// still decide
}

type TimestampMetaFeature struct {
	BaseID    uint `gorm:"not null"`
	ColumnNum int  `gorm:"not null"`
	// still decide
}

type SelectManyMetaFeature struct {
	BaseID    uint        `gorm:"not null"`
	ColumnNum int         `gorm:"not null"`
	Order     int         `gorm:"not null"`
	Value     string      `gorm:"not null"`
	Quality   QualityEnum `gorm:"not null"`
}
