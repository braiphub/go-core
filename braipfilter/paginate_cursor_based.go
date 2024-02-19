package braipfilter

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"time"

	"github.com/AlekSi/pointer"
	"github.com/pkg/errors"
	"gorm.io/gorm"
)

type PaginateCursor struct {
	NextPage     *string
	PreviousPage *string
}

type Cursor struct {
	Value     interface{} `json:"value"`
	Direction Direction   `json:"direction"`
}

type ItemID struct {
	ID uint
}

type Direction string

const (
	Forward  Direction = "forward"
	Backward Direction = "backward"
)

type CursorOrder string

const (
	OrderDESC CursorOrder = "DESC"
	OrderASC  CursorOrder = "ASC"
)

//nolint:funlen,cyclop
func (f *DBFilter) PaginateCursor(
	filterStruct interface{},
	tx *gorm.DB,
	items interface{},
	cursorField string,
	defaultOrder CursorOrder,
) (*PaginateCursor, error) {
	var paginate PaginateCursor
	var identifiers []interface{}
	var thisPageCursorValue interface{}
	var lastGotValue interface{}

	runningDirection := Forward

	config := f.paginateConfig(filterStruct)

	thisPageCursor, err := paginate.getThisPageCursor(config)
	if err != nil {
		return nil, errors.Wrap(err, "get this-page-cursor")
	}
	if thisPageCursor != nil {
		thisPageCursorValue = thisPageCursor.Value
		runningDirection = thisPageCursor.Direction
	}

	// search for identifiers
	if err := tx.Transaction(func(tx *gorm.DB) error {
		// current cursor applyment
		if thisPageCursor != nil {
			operator := paginate.getOperator(defaultOrder, runningDirection)

			tx = tx.Where(
				fmt.Sprintf("%s %s ?", cursorField, operator),
				thisPageCursorValue,
			)
		}

		// get ids
		orderBy := fmt.Sprintf("%s %s", cursorField, paginate.getOrder(defaultOrder, runningDirection))

		rows, err := tx.Select(cursorField).Order(orderBy).Rows() //nolint:rowserrcheck
		if err != nil {
			return errors.Wrap(err, "get rows")
		}
		defer rows.Close()

		for rows.Next() {
			var identifier interface{}

			if err := rows.Scan(&identifier); err != nil {
				return errors.Wrap(err, "scan row")
			}

			identifiers = append(identifiers, identifier)
			if len(identifiers) >= config.perPage {
				break
			}
		}

		// fill paginate cursors
		var hasMore bool
		var nextRowValue interface{}
		if rows.Next() {
			hasMore = true
			if err := rows.Scan(&nextRowValue); err != nil {
				return errors.Wrap(err, "scan row")
			}
		}

		if len(identifiers) > 0 {
			lastGotValue = identifiers[len(identifiers)-1]
		}
		if err := paginate.fillNextPage(runningDirection, lastGotValue, thisPageCursorValue, hasMore); err != nil {
			return errors.Wrap(err, "fill next page cursor")
		}

		if err := paginate.fillPrevPage(runningDirection, thisPageCursorValue, nextRowValue); err != nil {
			return errors.Wrap(err, "fill previous page cursor")
		}

		return nil
	}); err != nil {
		return nil, errors.Wrap(err, "transaction")
	}

	//nolint:asasalint
	// get model items
	if err := tx.Transaction(func(tx *gorm.DB) error {
		result := tx.
			Where(cursorField+" in (?)", identifiers).
			Order(cursorField + " " + string(defaultOrder)).
			Find(items)
		if result.Error != nil {
			return errors.Wrap(result.Error, "find items")
		}

		return nil
	}); err != nil {
		return nil, errors.Wrap(err, "transaction")
	}

	return &paginate, nil
}

func (p *PaginateCursor) getOrder(
	defaultOrder CursorOrder,
	direction Direction,
) CursorOrder {
	order := defaultOrder

	if direction == Backward {
		switch order {
		case OrderDESC:
			order = OrderASC
		case OrderASC:
			order = OrderDESC
		}
	}

	return order
}

func (p *PaginateCursor) getOperator(
	defaultOrder CursorOrder,
	direction Direction,
) string {
	switch {
	case defaultOrder == OrderDESC && direction == Forward:
		return "<"
	case defaultOrder == OrderDESC && direction == Backward:
		return ">="
	case defaultOrder == OrderASC && direction == Forward:
		return ">"
	case defaultOrder == OrderASC && direction == Backward:
		return "<="
	}

	return ""
}

func (p *PaginateCursor) getThisPageCursor(config paginateConfig) (*Cursor, error) {
	if config.cursorPage == nil {
		return nil, nil //nolint: nilnil
	}

	var cursor Cursor

	buf, err := base64.StdEncoding.DecodeString(*config.cursorPage)
	if err != nil {
		return nil, errors.Wrap(err, "decode current cursor")
	}

	if err := json.Unmarshal(buf, &cursor); err != nil {
		return nil, errors.Wrap(err, "unmarshal current cursor")
	}

	return &cursor, nil
}

func (p *PaginateCursor) fillNextPage(
	runningDirection Direction,
	lastGotValue interface{},
	thisPageCursorValue interface{},
	hasMore bool,
) error {
	cursor := Cursor{
		Value:     nil,
		Direction: Forward,
	}

	switch runningDirection {
	case Forward:
		if !hasMore {
			return nil
		}
		cursor.Value = lastGotValue
	case Backward:
		cursor.Value = thisPageCursorValue
	}

	if cursor.Value == nil {
		return nil
	}

	if timestamp, ok := cursor.Value.(time.Time); ok {
		cursor.Value = timestamp.Format(time.RFC3339Nano)
	}

	buf, err := json.Marshal(cursor)
	if err != nil {
		return errors.Wrap(err, "marshal cursor")
	}

	p.NextPage = pointer.ToString(base64.StdEncoding.EncodeToString(buf))

	return nil
}

func (p *PaginateCursor) fillPrevPage(runningDirection Direction, thisPageCursorValue, nextRowValue interface{}) error {
	cursor := Cursor{
		Value:     nil,
		Direction: Backward,
	}

	switch runningDirection {
	case Forward:
		cursor.Value = thisPageCursorValue
	case Backward:
		cursor.Value = nextRowValue
	}

	if cursor.Value == nil {
		return nil
	}

	if timestamp, ok := cursor.Value.(time.Time); ok {
		cursor.Value = timestamp.Format(time.RFC3339Nano)
	}

	buf, err := json.Marshal(cursor)
	if err != nil {
		return errors.Wrap(err, "marshal cursor")
	}

	p.PreviousPage = pointer.ToString(base64.StdEncoding.EncodeToString(buf))

	return nil
}
