package admin

import (
	"errors"
	"fmt"
	"html/template"
	"reflect"

	"github.com/jinzhu/gorm"
	"github.com/qor/qor"
	"github.com/qor/qor/resource"
	"github.com/qor/qor/utils"
)

// SelectManyConfig meta configuration used for select many
type SelectManyConfig struct {
	Collection               interface{} // []string, [][]string, func(interface{}, *qor.Context) [][]string, func(interface{}, *admin.Context) [][]string
	DefaultCreating          bool
	Placeholder              string
	SelectionTemplate        string
	SelectMode               string // select, select_async, bottom_sheet
	Select2ResultTemplate    template.JS
	Select2SelectionTemplate template.JS
	ForSerializedObject      bool
	RemoteDataResource       *Resource
	RemoteDataHasImage       bool
	PrimaryField             string
	SelectOneConfig
}

// GetTemplate get template for selection template
func (selectManyConfig SelectManyConfig) GetTemplate(context *Context, metaType string) ([]byte, error) {
	if metaType == "form" && selectManyConfig.SelectionTemplate != "" {
		return context.Asset(selectManyConfig.SelectionTemplate)
	}
	return nil, errors.New("not implemented")
}

// ConfigureQorMeta configure select many meta
func (selectManyConfig *SelectManyConfig) ConfigureQorMeta(metaor resource.Metaor) {
	if meta, ok := metaor.(*Meta); ok {
		selectManyConfig.SelectOneConfig.Collection = selectManyConfig.Collection
		selectManyConfig.SelectOneConfig.SelectMode = selectManyConfig.SelectMode
		selectManyConfig.SelectOneConfig.DefaultCreating = selectManyConfig.DefaultCreating
		selectManyConfig.SelectOneConfig.Placeholder = selectManyConfig.Placeholder
		selectManyConfig.SelectOneConfig.RemoteDataResource = selectManyConfig.RemoteDataResource
		selectManyConfig.SelectOneConfig.PrimaryField = selectManyConfig.PrimaryField

		selectManyConfig.SelectOneConfig.ConfigureQorMeta(meta)

		selectManyConfig.RemoteDataResource = selectManyConfig.SelectOneConfig.RemoteDataResource
		selectManyConfig.SelectMode = selectManyConfig.SelectOneConfig.SelectMode
		selectManyConfig.DefaultCreating = selectManyConfig.SelectOneConfig.DefaultCreating
		selectManyConfig.PrimaryField = selectManyConfig.SelectOneConfig.PrimaryField
		meta.Type = "select_many"

		// Set FormattedValuer
		if meta.FormattedValuer == nil {
			meta.SetFormattedValuer(func(record interface{}, context *qor.Context) interface{} {
				reflectValues := reflect.Indirect(reflect.ValueOf(meta.GetValuer()(record, context)))
				var results []string
				if reflectValues.IsValid() {
					for i := 0; i < reflectValues.Len(); i++ {
						results = append(results, utils.Stringify(reflectValues.Index(i).Interface()))
					}
				}
				return results
			})
		}
	}
}

// ConfigureQORAdminFilter configure admin filter
func (selectManyConfig *SelectManyConfig) ConfigureQORAdminFilter(filter *Filter) {
	var structField *gorm.StructField

	if field, ok := filter.Resource.GetAdmin().DB.NewScope(filter.Resource.Value).FieldByName(filter.Name); ok {
		structField = field.StructField
	}

	selectManyConfig.SelectOneConfig.Collection = selectManyConfig.Collection
	selectManyConfig.SelectOneConfig.SelectMode = selectManyConfig.SelectMode
	selectManyConfig.SelectOneConfig.RemoteDataResource = selectManyConfig.RemoteDataResource
	selectManyConfig.SelectOneConfig.PrimaryField = selectManyConfig.PrimaryField
	selectManyConfig.prepareDataSource(structField, filter.Resource, "!remote_data_filter")

	filter.Operations = []string{"In"}
	filter.Type = "select_many"
}

// FilterValue filter value
func (selectManyConfig *SelectManyConfig) FilterValue(filter *Filter, context *Context) interface{} {
	var (
		prefix  = fmt.Sprintf("filters[%v].", filter.Name)
		keyword interface{}
	)

	if metaValues, err := resource.ConvertFormToMetaValues(context.Request, []resource.Metaor{}, prefix); err == nil {
		if metaValue := metaValues.Get("Value"); metaValue != nil {
			if arr, ok := metaValue.Value.([]string); ok {
				keyword = arr
			} else {
				keyword = utils.ToString(metaValue.Value)
			}
		}
	}

	if keyword != nil && selectManyConfig.RemoteDataResource != nil {
		result := selectManyConfig.RemoteDataResource.NewSlice()
		clone := context.Clone()
		var primaryQuerySQL string
		for _, field := range selectManyConfig.RemoteDataResource.PrimaryFields {
			if filter.Name == field.DBName {
				primaryQuerySQL = fmt.Sprintf("%v in (?)", field.DBName)
			}
		}
		if primaryQuerySQL == "" {
			primaryQuerySQL = "id IN (?)"
		}

		clone.DB = clone.DB.Where(primaryQuerySQL, keyword)
		if selectManyConfig.RemoteDataResource.CallFindMany(result, clone) == nil {
			return result
		}
	}

	return keyword
}
