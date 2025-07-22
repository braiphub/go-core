package queue

import (
	"github.com/pkg/errors"
	"gorm.io/gorm"
	"reflect"
)

const (
	fallbackProducerTableName = "fallback_producer_queue_messages"
)

type GormFallbackProducerInterface interface {
	TableName() string
}

var _ GormFallbackProducerInterface = &GormFallbackProducerModel{}

type GormFallbackProducerModel struct {
	ID         uint `gorm:"primary_key"`
	RoutingKey string
	Headers    []byte
	Body       []byte
}

func (model GormFallbackProducerModel) TableName() string {
	return fallbackProducerTableName
}

type GormFallback struct {
	produceModel GormFallbackProducerInterface
	database     *gorm.DB
}

func NewGormFallBack(
	database *gorm.DB,
	produceModel GormFallbackProducerInterface,
) (*GormFallback, error) {
	if err := validateProducerTableColumns(produceModel); err != nil {
		return nil, errors.Wrap(err, "validate")
	}

	return &GormFallback{
		database:     database,
		produceModel: produceModel,
	}, nil
}

func validateProducerTableColumns(model GormFallbackProducerInterface) error {
	if model.TableName() != fallbackProducerTableName {
		return errors.New("produce table name must be: " + fallbackProducerTableName)
	}

	requiredFields := map[string]reflect.Kind{
		"ID":         reflect.Uint,
		"RoutingKey": reflect.String,
		"Headers":    reflect.Slice,
		"Body":       reflect.Slice,
	}

	for field, kind := range requiredFields {
		hasField := hasFieldWithType(model, field, kind)
		if !hasField {
			return errors.New("field " + field + " not found in produce model")
		}
	}

	return nil
}

func hasFieldWithType(model any, fieldName string, expectedType reflect.Kind) bool {
	t := reflect.TypeOf(model)

	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	if t.Kind() != reflect.Struct {
		return false
	}

	field, ok := t.FieldByName(fieldName)
	if !ok {
		return false
	}

	return field.Type.Kind() == expectedType
}
