package braipfilter

import (
	"reflect"
	"strings"

	"github.com/pkg/errors"
	"gorm.io/gorm"
)

type PaginatePageBased struct {
	Total        int
	Page         int
	PerPage      int
	NextPage     *int
	PreviousPage *int
	TotalPages   int
	To           int
}
type paginateConfig struct {
	page       int
	perPage    int
	cursorPage *string
	orderBy    string
}

func (f *DBFilter) PaginatePageBased(
	filterStruct interface{},
	tx *gorm.DB,
	items interface{},
) (*PaginatePageBased, error) {
	var paginate PaginatePageBased
	var count, filteredCount int64

	config := f.paginateConfig(filterStruct)

	// db count + db find
	if err := tx.Transaction(func(tx *gorm.DB) error {
		result := tx.
			Count(&count).
			Offset(config.offset()).
			Limit(config.perPage).
			Order(config.orderBy).
			Find(items)
		if result.Error != nil {
			return errors.Wrap(result.Error, "database find")
		}

		filteredCount = result.RowsAffected

		return nil
	}); err != nil {
		return nil, errors.Wrap(err, "transaction")
	}

	// fill paginate data
	paginate.fill(config.page, config.perPage, config.offset(), int(count), int(filteredCount))

	return &paginate, nil
}

//nolint:cyclop
func (f *DBFilter) paginateConfig(filterStruct interface{}) paginateConfig {
	var page, perPage int
	var cursorPage *string
	var orderBy []string

	reflectValue := reflect.Indirect(reflect.ValueOf(filterStruct))

	// page assimilate
	pageField := reflectValue.FieldByName("Page")
	if pageField.IsValid() {
		if i, ok := pageField.Interface().(int); ok {
			page = i
		}
	}
	if page <= 0 {
		page = 1
	}

	// per-page assimilate
	perPageField := reflectValue.FieldByName("PerPage")
	if perPageField.IsValid() {
		if i, ok := perPageField.Interface().(int); ok {
			perPage = i
		}
	}
	if perPage == 0 {
		perPage = f.defaultValues.perPage
	}

	// cursor page assimilate
	cursorPageField := reflectValue.FieldByName("Cursor")
	if cursorPageField.IsValid() {
		if cursor, ok := cursorPageField.Interface().(*string); ok && cursor != nil && *cursor != "" {
			cursorPage = cursor
		}
	}

	// order-by assimilate
	orderByField := reflectValue.FieldByName("OrderBy")
	if orderByField.IsValid() {
		if strSlice, ok := orderByField.Interface().([]string); ok {
			orderBy = strSlice
		}
	}
	for i, v := range orderBy {
		orderBy[i] = strings.ReplaceAll(v, ".", " ")
	}

	return paginateConfig{
		page:       page,
		perPage:    perPage,
		cursorPage: cursorPage,
		orderBy:    strings.Join(orderBy, ","),
	}
}

func (c *paginateConfig) offset() int {
	offset := (c.page - 1) * c.perPage

	return offset
}

func (p *PaginatePageBased) fill(page, perPage, offset, count, filteredCount int) {
	p.Page = page
	p.PerPage = perPage
	p.Total = count
	p.TotalPages = totalPages(count, perPage)
	p.To = offset + filteredCount

	// next page and previous page
	if page > 1 {
		previous := page - 1
		p.PreviousPage = &previous
	}
	if count > page*perPage {
		next := page + 1
		p.NextPage = &next
	}
}

func totalPages(count, perPage int) int {
	if perPage == 0 {
		return 0
	}

	rest := count % perPage

	pages := count / perPage
	if rest > 0 {
		pages++
	}

	return pages
}
