package fxsqls

import (
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/tacjlee/common-sdk/packages/fxmodels"
	"gorm.io/gorm"
	"reflect"
)

func FindFirst[T any](db *gorm.DB, fieldName string, fieldValue any) (T, error) {
	var result T
	var params map[string]any
	params = make(map[string]any)
	params[fieldName] = fieldValue
	query := db.Where(params)
	err := query.First(&result).Error
	if err != nil {
		return result, err
	}
	return result, nil
}

func FindOptionalObjectById[T any](db *gorm.DB, id interface{}) (fxmodels.Optional[T], error) {
	var results T
	parsedUUID, err := uuid.Parse(id.(string))
	if err != nil {
		return fxmodels.Optional[T]{Value: nil}, nil
	}
	if err := db.First(&results, parsedUUID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fxmodels.Optional[T]{Value: nil}, nil
		}
		return fxmodels.Optional[T]{Value: nil}, err
	}
	return fxmodels.Optional[T]{Value: &results}, nil
}

func ExecuteModelList[T any](db *gorm.DB, query string, params ...any) ([]T, error) {
	var results []T
	tx := db.Raw(query, params...).Scan(&results)
	if tx.Error != nil {
		if tx.Error == gorm.ErrRecordNotFound {
			return results, nil
		}
		return nil, fmt.Errorf("error executing query: %w", tx.Error)
	}
	return results, nil
}

func ExecuteModelObject[T any](db *gorm.DB, query string, params ...any) (T, error) {
	var result T
	if err := db.Raw(query, params...).First(&result).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return result, nil
		}
		// If there is an error, return nil
		return result, err
	}
	return result, nil
}

func ExecuteOptionalObject[T any](db *gorm.DB, query string, params ...any) (fxmodels.Optional[T], error) {
	var result T
	if err := db.Raw(query, params...).First(&result).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return fxmodels.Optional[T]{Value: nil}, nil
		}
		return fxmodels.Optional[T]{Value: nil}, err
	}
	return fxmodels.Optional[T]{Value: &result}, nil
}

func IsEmpty(value interface{}) bool {
	return reflect.DeepEqual(value, reflect.Zero(reflect.TypeOf(value)).Interface())
}

func DeleteAll[T any](db *gorm.DB, models []T) (int64, error) {
	var rowsEffected int64
	for _, model := range models {
		result := db.Delete(model)
		if result.Error != nil {
			return 0, fmt.Errorf("failed to delete model %v: %w", model, result.Error)
		}
		rowsEffected += result.RowsAffected
	}
	return rowsEffected, nil
}
