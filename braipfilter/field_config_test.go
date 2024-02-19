package braipfilter

import (
	"testing"

	"github.com/AlekSi/pointer"
	"github.com/stretchr/testify/assert"
)

func Test_scanFilterConfigFromStruct(t *testing.T) {
	type myStruct struct {
		Page             int      ``
		PerPage          *int     `filter:"default:15"`
		Ignore           *string  `filter:"-"`
		ID               *uint    ``
		ReferenceRenamed *string  `filter:"field:reference"`
		ReferenceLIKE    *string  `filter:"field:reference;default:buy"`
		Amount           *int     ``
		AmountGTE        *int     `filter:"default:-1"`
		StatusLTE        *int     `filter:"field:status;comparator=eq"`
		NonPointer       int      ``
		DefTrue          *bool    `filter:"default:true"`
		OrderBy          []string ``
	}

	expect := []filterConfig{
		{
			Index:        3,
			StructField:  "ID",
			DBField:      "id",
			Comparator:   comparatorEQ,
			DefaultValue: nil,
		},
		{
			Index:        4,
			StructField:  "ReferenceRenamed",
			DBField:      "reference",
			Comparator:   comparatorEQ,
			DefaultValue: nil,
		},
		{
			Index:        5,
			StructField:  "ReferenceLIKE",
			DBField:      "reference",
			Comparator:   comparatorLIKE,
			DefaultValue: pointer.ToString("buy"),
		},
		{
			Index:        6,
			StructField:  "Amount",
			DBField:      "amount",
			Comparator:   comparatorEQ,
			DefaultValue: nil,
		},
		{
			Index:        7,
			StructField:  "AmountGTE",
			DBField:      "amount",
			Comparator:   comparatorGTE,
			DefaultValue: pointer.ToInt(-1),
		},
		{
			Index:        8,
			StructField:  "StatusLTE",
			DBField:      "status",
			Comparator:   comparatorLTE,
			DefaultValue: nil,
		},
		{
			Index:        9,
			StructField:  "NonPointer",
			DBField:      "non_pointer",
			Comparator:   comparatorEQ,
			DefaultValue: nil,
		},
		{
			Index:        10,
			StructField:  "DefTrue",
			DBField:      "def_true",
			Comparator:   comparatorEQ,
			DefaultValue: pointer.ToBool(true),
		},
	}

	got := scanFilterConfigFromStruct(myStruct{})
	assert.Equal(t, expect, got)
}
