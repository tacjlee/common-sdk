package fxsql

import (
	"reflect"

	"github.com/tacjlee/common-sdk/packages/fxmodel"
	"github.com/tacjlee/common-sdk/packages/fxrepository"
	"gorm.io/gorm"
)

// The typed query logic now lives on *fxrepository.GenericRepository as generic
// methods (Go 1.27). These free functions are retained as thin shims so existing
// callers keep compiling; new code should prefer fxrepository.New(db).<Method>[T].

func FindFirst[T any](db *gorm.DB, fieldName string, fieldValue any) (T, error) {
	return fxrepository.New(db).FindFirst[T](fieldName, fieldValue)
}

func FindOptionalObjectById[T any](db *gorm.DB, id interface{}) (*fxmodel.Optional[T], error) {
	return fxrepository.New(db).FindOptionalObjectById[T](id)
}

func ExecuteModelList[T any](db *gorm.DB, query string, params ...any) ([]T, error) {
	return fxrepository.New(db).ExecuteModelList[T](query, params...)
}

func ExecuteModelObject[T any](db *gorm.DB, query string, params ...any) (T, error) {
	return fxrepository.New(db).ExecuteModelObject[T](query, params...)
}

func ExecuteOptionalObject[T any](db *gorm.DB, query string, params ...any) (fxmodel.Optional[T], error) {
	return fxrepository.New(db).ExecuteOptionalObject[T](query, params...)
}

func DeleteAll[T any](db *gorm.DB, models []T) (int64, error) {
	return fxrepository.New(db).DeleteModels[T](models)
}

func IsEmpty(value interface{}) bool {
	return reflect.DeepEqual(value, reflect.Zero(reflect.TypeOf(value)).Interface())
}
