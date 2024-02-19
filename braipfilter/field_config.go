package braipfilter

import (
	"reflect"
	"strconv"
	"strings"

	"golang.org/x/exp/slices"
)

type filterConfig struct {
	Index        int
	StructField  string
	DBField      string
	Comparator   Comparator
	DefaultValue interface{}
	Relationship *string
}

type Comparator int

const (
	comparatorUnset Comparator = iota
	comparatorEQ
	comparatorLIKE
	comparatorLT
	comparatorLTE
	comparatorGT
	comparatorGTE
)

func (f *filterConfig) toSQL() string {
	var comparator string

	value := "?"

	switch f.Comparator {
	case comparatorUnset:
		comparator = "="
	case comparatorEQ:
		comparator = "="
	case comparatorLIKE:
		comparator = "ILIKE"
		value = "?"
	case comparatorLT:
		comparator = "<"
	case comparatorLTE:
		comparator = "<="
	case comparatorGT:
		comparator = ">"
	case comparatorGTE:
		comparator = ">="
	}

	return f.DBField + " " + comparator + " " + value
}

func scanFilterConfigFromStruct(input interface{}) []filterConfig {
	var filters []filterConfig

	st := reflect.TypeOf(input)

	for i := 0; i < st.NumField(); i++ {
		field := st.Field(i)
		tags := getTags(field)

		if shouldIgnoreStructField(field) || !isFilterField(field.Name) {
			continue
		}
		if field.Anonymous {
			continue
		}

		fieldName := toSnakeCase(fieldNameWithoutComparator(field.Name))
		comparator := comparatorEQ
		if name := fieldNameFromTag(tags); name != "" {
			fieldName = name
		}
		if c := comparatorFromFieldName(field.Name); c != comparatorUnset {
			comparator = c
		}
		if tag := comparatorFromTag(tags); tag != comparatorUnset {
			comparator = tag
		}

		var relationship *string
		if val, ok := tags["relationship"]; ok {
			relationship = &val
			fieldName = *relationship + "." + fieldName
		}

		filters = append(filters, filterConfig{
			Index:        field.Index[0],
			StructField:  field.Name,
			DBField:      fieldName,
			Comparator:   comparator,
			DefaultValue: getDefaultValue(field, tags),
			Relationship: relationship,
		})
	}

	return filters
}

func isFilterField(fieldName string) bool {
	ignoreFieldNames := []string{"Page", "PerPage", "OrderBy", "Cursor"}

	return !slices.Contains(ignoreFieldNames, fieldName)
}

func fieldNameWithoutComparator(name string) string {
	trimSuffixes := []string{"EQ", "LIKE", "LT", "LTE", "GT", "GTE"}

	for _, suffix := range trimSuffixes {
		if strings.HasSuffix(name, suffix) {
			return strings.TrimSuffix(name, suffix)
		}
	}

	return name
}

func fieldNameFromTag(tags map[string]string) string {
	if tagValue, ok := tags["field"]; ok {
		return tagValue
	}

	return ""
}

func comparatorFromFieldName(name string) Comparator {
	trimSuffixes := map[string]Comparator{
		"EQ":   comparatorEQ,
		"LIKE": comparatorLIKE,
		"LT":   comparatorLT,
		"LTE":  comparatorLTE,
		"GT":   comparatorGT,
		"GTE":  comparatorGTE,
	}

	for suffix, comparator := range trimSuffixes {
		if strings.HasSuffix(name, suffix) {
			return comparator
		}
	}

	return comparatorUnset
}

func comparatorFromTag(tags map[string]string) Comparator {
	comparatorMap := map[string]Comparator{
		"EQ":   comparatorEQ,
		"LIKE": comparatorLIKE,
		"LT":   comparatorLT,
		"LTE":  comparatorLTE,
		"GT":   comparatorGT,
		"GTE":  comparatorGTE,
	}

	if tagValue, ok := tags["comparator"]; ok {
		return comparatorMap[tagValue]
	}

	return comparatorUnset
}

func getDefaultValue(field reflect.StructField, tags map[string]string) interface{} {
	if tagValue, ok := tags["default"]; ok {
		switch field.Type.String() {
		case "*string":
			return &tagValue

		case "*int":
			def, _ := strconv.Atoi(tagValue)

			return &def

		case "*bool":
			def, _ := strconv.ParseBool(tagValue)

			return &def
		}
	}

	return nil
}
