package braipfilter

import (
	"testing"

	"github.com/AlekSi/pointer"
	"github.com/stretchr/testify/assert"
)

func Test_getFiltersFromStruct(t *testing.T) {
	data := fixtureMySearchStruct()

	expected := []fieldFilter{
		{SQL: "id = ?", Value: pointer.ToUint(1)},
		{SQL: "reference = ?", Value: pointer.ToString("buyrley6")},
		{SQL: "reference ILIKE ?", Value: pointer.ToString("%buy%")},
		{SQL: "amount = ?", Value: pointer.ToInt(46000)},
		{SQL: "amount >= ?", Value: pointer.ToInt(45999)},
		{SQL: "status <= ?", Value: pointer.ToInt(2)},
		{SQL: "non_pointer_integer = ?", Value: 1},
		{SQL: "def_true = ?", Value: pointer.ToBool(true)},
		{Relationship: pointer.ToString("customers"), SQL: "customers.name = ?", Value: pointer.ToString("customer name")},
	}

	configs := scanFilterConfigFromStruct(data)

	got := getFiltersFromStruct(data, configs)
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
