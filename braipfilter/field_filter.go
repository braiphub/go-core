package braipfilter

import (
	"reflect"

	"github.com/AlekSi/pointer"
)

type fieldFilter struct {
	Relationship *string
	SQL          string
	Value        interface{}
}

func getFiltersFromStruct(input interface{}, configs []filterConfig) []fieldFilter {
	var filters []fieldFilter

	for _, config := range configs {
		value := getFilterValue(input, config.StructField, config.Comparator, config.DefaultValue)
		if value == nil {
			continue
		}

		filters = append(filters, fieldFilter{
			Relationship: config.Relationship,
			SQL:          config.toSQL(),
			Value:        value,
		})
	}

	return filters
}

func getFilterValue(input interface{}, fieldName string, comparator Comparator, defaultValue interface{}) interface{} {
	var value interface{}

	rv := reflect.Indirect(reflect.ValueOf(input))

	field := rv.FieldByName(fieldName)
	value = field.Interface()

	if field.Kind() == reflect.Ptr && reflect.ValueOf(value).IsNil() {
		value = defaultValue
	}

	// add percentage to like: LIKE "%search%"
	if comparator == comparatorLIKE {
		switch originalValue := value.(type) {
		case *string:
			value = pointer.ToString("%" + *originalValue + "%")
		case string:
			value = "%" + originalValue + "%"
		}
	}

	return value
}
