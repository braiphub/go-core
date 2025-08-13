package braipfilter

import (
	"testing"

	"github.com/AlekSi/pointer"
	"github.com/stretchr/testify/assert"
)

func Test_getFiltersFromStruct(t *testing.T) {
	data := fixtureMySearchStruct()

	expected := []fieldFilter{
		{SQL: "table_name.id = ?", Value: pointer.ToUint(1)},
		{SQL: "table_name.reference = ?", Value: pointer.ToString("buyrley6")},
		{SQL: "table_name.reference ILIKE ?", Value: pointer.ToString("%buy%")},
		{SQL: "table_name.amount = ?", Value: pointer.ToInt(46000)},
		{SQL: "table_name.amount >= ?", Value: pointer.ToInt(45999)},
		{SQL: "table_name.status <= ?", Value: pointer.ToInt(2)},
		{SQL: "table_name.non_pointer_integer = ?", Value: 1},
		{SQL: "table_name.def_true = ?", Value: pointer.ToBool(true)},
		{Relationship: pointer.ToString("customers"), SQL: "customers.name = ?", Value: pointer.ToString("customer name")},
	}

	configs := scanFilterConfigFromStruct(data)

	got := getFiltersFromStruct(data, configs, "table_name")
	assert.Equal(t, expected, got)
}

func Test_getFilterValue(t *testing.T) {
	data := fixtureMySearchStruct()

	got := getFilterValue(data, "ID", comparatorEQ, nil)
	assert.Equal(t, data.ID, got)

	got = getFilterValue(data, "DefTrue", comparatorEQ, pointer.ToBool(true))
	assert.Equal(t, pointer.ToBool(true), got)

	got = getFilterValue(data, "DefTrue", comparatorEQ, pointer.ToBool(false))
	assert.Equal(t, pointer.ToBool(false), got)
}
