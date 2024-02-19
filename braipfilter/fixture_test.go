package braipfilter

import (
	"github.com/AlekSi/pointer"
)

type mySearchStruct struct {
	Page              int     ``
	PerPage           *int    `filter:"default:15"`
	Ignore            *string `filter:"-"`
	ID                *uint   ``
	ReferenceRenamed  *string `filter:"field:reference"`
	ReferenceLIKE     *string `filter:"field:reference;default:buy"`
	Amount            *int    ``
	AmountGTE         *int    ``
	StatusLTE         *int    `filter:"field:status;comparator=eq"`
	StatusUnfilled    *int    `filter:"field:status;comparator=eq"`
	NonPointerInteger int     ``
	DefTrue           *bool   `filter:"default:true"`
	CustomerName      *string `filter:"relationship:customers;field:name"`
}

func fixtureMySearchStruct() mySearchStruct {
	return mySearchStruct{
		Ignore:            pointer.ToString("ignored field"),
		ID:                pointer.ToUint(1),
		ReferenceRenamed:  pointer.ToString("buyrley6"),
		Amount:            pointer.ToInt(46000),
		AmountGTE:         pointer.ToInt(45999),
		StatusLTE:         pointer.ToInt(2), // forced eq
		PerPage:           nil,
		NonPointerInteger: 1,
		CustomerName:      pointer.ToString("customer name"),
	}
}
