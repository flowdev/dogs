package mygorm

import (
	"github.com/jinzhu/gorm"
)

type TypeEnum struct {
	slug string
}

func (te TypeEnum) String() string {
	return te.slug
}

var (
	TypeUnknown    = TypeEnum{""}
	TypeString     = TypeEnum{"string"}
	TypeText       = TypeEnum{"text"}
	TypeCheckbox   = TypeEnum{"checkbox"}
	TypeInt        = TypeEnum{"integer"}
	TypeFloat      = TypeEnum{"floating point"}
	TypeDate       = TypeEnum{"date"}
	TypeDateTime   = TypeEnum{"timestamp"}
	TypeSelectOne  = TypeEnum{"select one"}
	TypeSelectMany = TypeEnum{"select many"}
)

var AllTypeEnums = []string{
	TypeUnknown.slug,
	TypeString.slug,
	TypeText.slug,
	TypeCheckbox.slug,
	TypeInt.slug,
	TypeFloat.slug,
	TypeDate.slug,
	TypeDateTime.slug,
	TypeSelectOne.slug,
	TypeSelectMany.slug,
}

func NewTypeEnum(s string) TypeEnum {
	switch s {
	case TypeString.slug:
		return TypeString
	case TypeText.slug:
		return TypeText
	case TypeCheckbox.slug:
		return TypeCheckbox
	case TypeInt.slug:
		return TypeInt
	case TypeFloat.slug:
		return TypeFloat
	case TypeDate.slug:
		return TypeDate
	case TypeDateTime.slug:
		return TypeDateTime
	case TypeSelectOne.slug:
		return TypeSelectOne
	case TypeSelectMany.slug:
		return TypeSelectMany
	}
	return TypeUnknown
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
