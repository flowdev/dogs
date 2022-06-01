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

type MetaType struct {
	gorm.Model
	Name  string
	Group string
	Type  TypeEnum
}

type MetaGroup struct {
	gorm.Model
	Name string
}
