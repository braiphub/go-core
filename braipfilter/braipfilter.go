package braipfilter

import (
	"fmt"
	"reflect"
	"strings"

	"gorm.io/gorm"
)

type DBFilter struct {
	defaultValues defaultValues
	cache         *filterCache
}

type defaultValues struct {
	perPage int
}

const perPageDefault = 10

func New() *DBFilter {
	return &DBFilter{
		defaultValues: defaultValues{
			perPage: perPageDefault,
		},
		cache: newCache(),
	}
}

func (f *DBFilter) FilterScope(filterStruct interface{}) func(db *gorm.DB) *gorm.DB {
	configs := f.filterConfigCached(filterStruct)
	structFilters := getFiltersFromStruct(filterStruct, configs)
	filters := filtersWithouRelationship(structFilters)
	relationshipFiltersGroupped := groupRelationshipFilters(structFilters)

	return func(db *gorm.DB) *gorm.DB {
		relatedTableName := f.getRelatedTableName(db)
		relatedPrefix := f.getRelatedPrefix(db)

		for _, filter := range filters {
			db = db.Where(filter.SQL, filter.Value)
		}

		for _, filter := range relationshipFiltersGroupped {
			sql := fmt.Sprintf(
				"JOIN %s ON %s.%s_id = %s.id",
				*filter.Relationship, // join
				*filter.Relationship,
				relatedPrefix,
				relatedTableName, // destination table
			)
			filter.SQL = fmt.Sprintf("%s.deleted_at IS NULL AND %s", *filter.Relationship, filter.SQL)
			db = db.Joins(sql).Where(filter.SQL, filter.Value.([]interface{})...) //nolint:forcetypeassert
		}

		return db
	}
}

func filtersWithouRelationship(filters []fieldFilter) []fieldFilter {
	newFiters := []fieldFilter{}
	for _, f := range filters {
		if f.Relationship != nil {
			continue
		}

		newFiters = append(newFiters, f)
	}

	return newFiters
}

func groupRelationshipFilters(filters []fieldFilter) []fieldFilter {
	filterMap := map[string]fieldFilter{}

	for _, filter := range filters {
		if filter.Relationship == nil {
			continue
		}

		existing, ok := filterMap[*filter.Relationship]
		if !ok {
			filterMap[*filter.Relationship] = fieldFilter{
				Relationship: filter.Relationship,
				SQL:          filter.SQL,
				Value:        []interface{}{filter.Value},
			}

			continue
		}

		existing.SQL += " AND " + filter.SQL
		existing.Value = append(existing.Value.([]interface{}), filter.Value) //nolint:forcetypeassert
		filterMap[*filter.Relationship] = existing
	}

	newFiters := []fieldFilter{}
	for _, f := range filterMap {
		newFiters = append(newFiters, f)
	}

	return newFiters
}

func (f *DBFilter) filterConfigCached(filterStruct interface{}) []filterConfig {
	structName := reflect.TypeOf(filterStruct).String()

	configs := f.cache.get(structName)
	if len(configs) == 0 {
		configs = scanFilterConfigFromStruct(filterStruct)
		f.cache.set(structName, configs)
	}

	return configs
}

func (f *DBFilter) getRelatedTableName(db *gorm.DB) string {
	modelName := f.getRelatedPrefix(db)
	tableName := db.NamingStrategy.TableName(modelName)

	return tableName
}

func (f *DBFilter) getRelatedPrefix(db *gorm.DB) string {
	relatedPrefix := reflect.TypeOf(db.Statement.Model).String()

	if strings.Contains(relatedPrefix, ".") {
		parts := strings.Split(relatedPrefix, ".")
		relatedPrefix = parts[len(parts)-1]
	}

	relatedPrefix = strings.ToLower(relatedPrefix)

	return relatedPrefix
}

func shouldIgnoreStructField(field reflect.StructField) bool {
	return field.Tag.Get("filter") == "-"
}

func getTags(field reflect.StructField) map[string]string {
	const keyValueLen = 2

	tags := map[string]string{}

	fullTag := field.Tag.Get("filter")

	tagList := strings.Split(fullTag, ";")
	for _, tagEncoded := range tagList {
		kv := strings.Split(tagEncoded, ":")
		key := kv[0]
		value := ""

		if len(kv) == keyValueLen {
			value = kv[1]
		}

		tags[key] = value
	}

	return tags
}
