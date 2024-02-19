
### REPOSITORY IMPLEMENT EXAMPLE:
```golang
type SampleRepo struct {
	db          *gorm.DB
	braipFilter *braipfilter.DBFilter
}

func New(db *gorm.DB) *SampleRepo {
	return &SampleRepo{
		db:          db,
		braipFilter: braipfilter.New(),
	}
}

func (r *repo) SearchCursorBased(filters entity.SearchRequest) (*entity.SearchResponse, error) {
	var response entity.SearchResponse

	//nolint:exhaustruct
	tx := r.db.WithContext(ctx).
		Model(&MyStruct{}).
		Preload(clause.Associations).
		Scopes(r.braipFilter.FilterScope(filters))

	cursor, err := r.braipFilter.PaginateCursor(
		filters,
		tx,
		&response.Items,
		"created_at",
		braipfilter.OrderDESC,
	)
	if err != nil {
		return nil, errors.Wrap(err, "paginate")
	}

	response.PaginateCursor = *cursor

	return &response, nil
}

func (r *repo) SearchPaginateBased(filters SearchRequest) (*SearchResponse, error) {
	var response SearchResponse

	//nolint:exhaustruct
	tx := r.db.WithContext(ctx).
		Model(&MyStruct{}).
		Preload(clause.Associations).
		Scopes(r.braipFilter.FilterScope(filters))

	paginate, err := r.braipFilter.PaginatePageBased(filters, tx, &response.Items)
	if err != nil {
		return nil, errors.Wrap(err, "paginate")
	}
	response.Paginate = *paginate

	return &response, nil
}
```


### REQUEST STRUCT:

* You need to implement  asdf

```golang
// Field annotations:
// [braipfilter] for this package;
// [query] are labstack/echo annotations.

type PaginateRequest struct {
	Page    int      `query:"page"`
	PerPage int      `query:"per_page"`
	OrderBy []string `query:"order_by"`
}

type SearchRequest struct {
	PaginateRequest
	ID               *uint          `query:"id"`
	Reference        *string        `query:"reference"`
	Currency         string         `query:"currency"`
	NameLIKE         *string        `query:"customer_name[like]"`
	EmailEQ          *string        `query:"customer_email"`
	Document         *string        `query:"customer_document"`
	CreatedAtGTE     *string        `query:"accepted_at[gte]"`
    CreatedAtLTE     *string        `query:"accepted_at[gte]"`
	AmountGT         *int           `query:"amount[gte]"`
	AmountLT         *int           `query:"amount[lte]"`
	CustomerName     *string        `braipfilter:"relationship:customers;field:name"` // join and filter customers.name relationship
}

// or
type SearchRequestWithAnnotations struct {
	Page              int      ``
	PerPage           *int     `filter:"default:15"` // if not specified, will be 15 records per page
    OrderBy           []string ``
	Ignore            *string  `filter:"-"` // will be ignored
	ReferenceForced   *string  `filter:"field:reference"` // reference = ?
	ReferenceLIKE     *string  `filter:"default:buy"` // reference LIKE "%buy%"
	Amount            *int     `` // amount = ?
	AmountGTE         *int     `` // amount >= ?
	StatusLTE         *int     `filter:"field:status;comparator=eq"` // status = ?
	NonPointerInteger int      `` // will filter as 0 or any entered number. non_pointer_integer = 0
	DefTrue           *bool    `filter:"default:true"` // default true. def_true = true
	CustomerName     *string   `filter:"relationship:customers;field:name"` // join and filter customers.name relationship
}

```


### RESPONSE STRUCT:
```golang
type SearchResponse struct {
	braipfilter.Paginate
	Items []MyStruct
}
```

