package braipfilter

import (
	"testing"
)

// import (
// 	"github.com/stretchr/testify/assert"
// 	"gorm.io/driver/sqlite"
// 	"gorm.io/gorm"
// 	"testing"
// 	"github.com/AlekSi/pointer"
// )
//
// type myEntity struct {
// 	gorm.Model
// 	Reference         string
// 	Amount            int
// 	Status            int
// 	NonPointerInteger int
// 	DefTrue           bool
// 	Customer          Customer
// }
//
// type Customer struct {
// 	gorm.Model
// 	MyEntityID uint
// 	Name       string
// }
//
// func Test_FilterScope(t *testing.T) {
// 	db, err := gorm.Open(sqlite.Open("file::memory:"), &gorm.Config{})
// 	assert.NoError(t, err)
// 	assert.NotNil(t, db)
// 	assert.NoError(t, db.AutoMigrate(
// 		&myEntity{},
// 		&Customer{},
// 	))
//
// 	// feed db
// 	orders := []myEntity{
// 		{
// 			Reference:         "buy-1",
// 			Amount:            100,
// 			Status:            2,
// 			NonPointerInteger: 0,
// 			DefTrue:           true,
// 			Customer: Customer{
// 				Name: "customer 1",
// 			},
// 		},
// 		{
// 			Reference:         "buy-2",
// 			Amount:            200,
// 			Status:            1,
// 			NonPointerInteger: 7,
// 			DefTrue:           false,
// 			Customer: Customer{
// 				Name: "customer 2",
// 			},
// 		},
// 	}
// 	assert.NoError(t, db.Create(orders).Error)
// 	assert.Equal(t, uint(1), orders[0].ID)
// 	assert.Equal(t, uint(2), orders[1].ID)
//
// 	dbFilter := New()
//
// 	// auto first
// 	err = db.Transaction(func(tx *gorm.DB) error {
// 		var items []myEntity
//
// 		search := mySearchStruct{}
// 		err = tx.Model(&myEntity{}).Scopes(dbFilter.FilterScope(search)).Find(&items).Error
// 		assert.NoError(t, err)
//
// 		assert.Len(t, items, 1)
// 		assert.Equal(t, uint(1), items[0].ID)
//
// 		return nil
// 	})
// 	assert.NoError(t, err)
//
// 	// exactly first
// 	err = db.Transaction(func(tx *gorm.DB) error {
// 		var items []myEntity
//
// 		search := mySearchStruct{
// 			ID:                pointer.ToUint(1),
// 			ReferenceRenamed:  pointer.ToString("buy-1"),
// 			ReferenceLIKE:     pointer.ToString("buy-1"),
// 			Amount:            pointer.ToInt(100),
// 			AmountGTE:         pointer.ToInt(99),
// 			StatusLTE:         pointer.ToInt(2),
// 			StatusUnfilled:    pointer.ToInt(2),
// 			NonPointerInteger: 0,
// 			DefTrue:           pointer.ToBool(true),
// 		}
// 		err = tx.Model(&myEntity{}).Scopes(dbFilter.FilterScope(search)).Find(&items).Error
// 		assert.NoError(t, err)
//
// 		assert.Len(t, items, 1)
// 		assert.Equal(t, uint(1), items[0].ID)
//
// 		return nil
// 	})
// 	assert.NoError(t, err)
//
// 	// exactly second
// 	err = db.Transaction(func(tx *gorm.DB) error {
// 		var items []myEntity
//
// 		search := mySearchStruct{
// 			ID:                pointer.ToUint(2),
// 			ReferenceRenamed:  pointer.ToString("buy-2"),
// 			ReferenceLIKE:     pointer.ToString("buy-2"),
// 			Amount:            pointer.ToInt(200),
// 			AmountGTE:         pointer.ToInt(199),
// 			StatusLTE:         pointer.ToInt(1),
// 			StatusUnfilled:    pointer.ToInt(1),
// 			NonPointerInteger: 7,
// 			DefTrue:           pointer.ToBool(false),
// 		}
// 		err = tx.Model(&myEntity{}).Scopes(dbFilter.FilterScope(search)).Find(&items).Error
// 		assert.NoError(t, err)
//
// 		assert.Len(t, items, 1)
// 		assert.Equal(t, uint(2), items[0].ID)
//
// 		return nil
// 	})
// 	assert.NoError(t, err)
// }

// func Test_FilterScope_Relationship(t *testing.T) {
// 	db, err := gorm.Open(sqlite.Open("file::memory:"), &gorm.Config{})
// 	assert.NoError(t, err)
// 	assert.NotNil(t, db)
// 	assert.NoError(t, db.AutoMigrate(
// 		&myEntity{},
// 		&Customer{},
// 	))

// 	// feed db
// 	orders := []myEntity{
// 		{
// 			Reference: "buy-1",
// 			Amount:    100,
// 			Status:    2,
// 			DefTrue:   true,
// 			Customer: Customer{
// 				Name: "customer 1",
// 			},
// 		},
// 		{
// 			Reference: "buy-2",
// 			Amount:    200,
// 			Status:    1,
// 			DefTrue:   true,
// 			Customer: Customer{
// 				Name: "customer 2",
// 			},
// 		},
// 	}
// 	assert.NoError(t, db.Create(orders).Error)
// 	assert.Equal(t, uint(1), orders[0].ID)
// 	assert.Equal(t, uint(2), orders[1].ID)

// 	dbFilter := New()

// 	// filter by relationship: not found
// 	db = db.Debug()
// 	err = db.Transaction(func(tx *gorm.DB) error {
// 		var items []myEntity

// 		search := mySearchStruct{
// 			ReferenceLIKE: pointer.ToString(""),
// 			CustomerName:  pointer.ToString("invalid"),
// 		}
// 		err = tx.Model(&myEntity{}).Scopes(dbFilter.FilterScope(search)).Find(&items).Error
// 		assert.NoError(t, err)

// 		assert.Len(t, items, 0)

// 		return nil
// 	})
// 	assert.NoError(t, err)

// 	// filter by relationship: ok
// 	err = db.Transaction(func(tx *gorm.DB) error {
// 		var items []myEntity

// 		search := mySearchStruct{
// 			ReferenceLIKE: pointer.ToString(""),
// 			CustomerName:  pointer.ToString("customer 2"),
// 		}
// 		err = tx.Model(&myEntity{}).Scopes(dbFilter.FilterScope(search)).Find(&items).Error
// 		assert.NoError(t, err)

// 		assert.Len(t, items, 1)
// 		assert.Equal(t, uint(2), items[0].ID)

// 		return nil
// 	})
// 	assert.NoError(t, err)
// }

func Benchmark_GetConfigsWithoutCache(b *testing.B) {
	type sample struct {
		Names []string
	}

	data := sample{
		Names: []string{"test1", "test2", "test3"},
	}

	for n := 0; n < b.N; n++ {
		scanFilterConfigFromStruct(data)
	}
}

func Benchmark_GetConfigsFromCache(b *testing.B) {
	type sample struct {
		Names []string
	}

	data := sample{
		Names: []string{"test1", "test2", "test3"},
	}

	dbFilter := New()

	for n := 0; n < b.N; n++ {
		dbFilter.filterConfigCached(data)
	}
}
