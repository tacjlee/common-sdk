package fxrepository

import (
	"database/sql"
	"errors"
	"fmt"
	"math"
	"reflect"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"github.com/tacjlee/common-sdk/packages/fxmodel"
	"github.com/tacjlee/common-sdk/packages/fxstring"
	"gorm.io/gorm"
)

// IGenericRepository is kept for backward compatibility with existing consumers
// that inject the repository as an interface. It only exposes the non-generic
// methods, because interface methods cannot declare type parameters. Typed
// (generic) operations live directly on *GenericRepository.
type IGenericRepository interface {
	GetDB() *gorm.DB
	ExecuteNonQuery(command string, params ...any) (int64, error)
	ExecuteJsonList(query string, params ...any) ([]map[string]any, error)
	ExecuteJsonPaging(query string, pageable fxmodel.Pageable, params ...any) (map[string]any, error)
	ExecuteKeyValueList(keyAlias string, valueAlias string, query string, params ...any) ([]map[string]any, error)
	ExecuteJsonObject(query string, params ...any) (map[string]any, error)
	ExecuteStringList(query string, params ...any) ([]string, error)
	ExecuteScalar(query string, params ...any) (any, error)
	ExecuteScalarAsBool(query string, params ...any) (bool, error)
	ExecuteScalarAsString(query string, params ...any) (string, error)
	ExecuteScalarAsLong(query string, params ...any) (int64, error)
	Create(value any) (any, error)
	Save(record any) (any, error)
	Delete(model any, conditions ...any) (int64, error)
	DeleteAll(models []any) (int64, error)
}

type GenericRepository struct {
	db *gorm.DB
}

// New builds a repository bound to db. Prefer this and use the concrete
// *GenericRepository so the generic (typed) methods are reachable.
func New(db *gorm.DB) *GenericRepository {
	return &GenericRepository{db: db}
}

// NewGenericRepository is retained for backward compatibility. It returns the
// interface, so callers keep the non-generic surface; switch to New to reach
// the typed methods.
func NewGenericRepository(db *gorm.DB) IGenericRepository {
	return New(db)
}

func (this *GenericRepository) GetDB() *gorm.DB {
	return this.db
}

// With returns a repository bound to a different *gorm.DB (e.g. a transaction
// handle), so typed methods can run inside db.Transaction(func(tx){...}).
func (this *GenericRepository) With(db *gorm.DB) *GenericRepository {
	return &GenericRepository{db: db}
}

// For update, delete command
func (this *GenericRepository) ExecuteNonQuery(query string, params ...any) (int64, error) {
	var result *gorm.DB
	if len(params) == 0 {
		result = this.db.Exec(query)
	} else {
		result = this.db.Exec(query, params...)
	}
	if result.Error != nil {
		return 0, result.Error
	}
	rowsAffected := result.RowsAffected
	return rowsAffected, nil
}

func (this *GenericRepository) ExecuteJsonList(query string, params ...any) ([]map[string]any, error) {
	var rows *sql.Rows
	var err error
	if len(params) == 0 {
		rows, err = this.db.Raw(query).Rows()
	} else {
		rows, err = this.db.Raw(query, params...).Rows()
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	// Prepare a slice to hold the results
	var result = make([]map[string]any, 0) //Empty slide
	columns, err := rows.Columns()
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		// Create a slice of interface{} to hold each column value
		columnsData := make([]interface{}, len(columns))
		columnPointers := make([]interface{}, len(columns))
		for i := range columnsData {
			columnPointers[i] = &columnsData[i]
		}
		// Scan the row into the slice
		if ex := rows.Scan(columnPointers...); ex != nil {
			return nil, ex
		}
		// Create a map to hold column names and their values
		rowMap := make(map[string]interface{})
		for i, col := range columns {
			jsonField := fxstring.ToJsonCase(col)
			// Convert []byte to string to avoid base64 encoding in JSON
			if b, ok := columnsData[i].([]byte); ok {
				rowMap[jsonField] = string(b)
			} else {
				rowMap[jsonField] = columnsData[i]
			}
		}
		// Add the map to the result slice
		result = append(result, rowMap)
	}
	return result, nil
}

func (this *GenericRepository) ExecuteJsonPaging(query string, pageable fxmodel.Pageable, params ...any) (map[string]any, error) {
	countingSql := this.buildCountingQuery(query)
	totalItems, err := this.ExecuteScalarAsLong(countingSql, params...)
	if err != nil {
		return nil, err
	}
	totalPages := int(math.Ceil(float64(totalItems) / float64(pageable.PageSize)))
	isLastPage := pageable.PageNumber >= totalPages
	result := make(map[string]any)
	result["totalItems"] = totalItems
	result["totalPages"] = totalPages
	result["pageSize"] = pageable.PageSize
	result["pageNumber"] = pageable.PageNumber
	result["items"] = make([]map[string]any, 0)
	result["isLastPage"] = isLastPage

	if totalItems == 0 || pageable.PageSize <= 0 {
		return result, nil
	}
	sortingClause := this.buildSortingClause(pageable.Order)
	limitingClause := this.buildLimitingClause(pageable)
	pagingSql := query + sortingClause + limitingClause

	items, errData := this.ExecuteJsonList(pagingSql, params...)
	if errData != nil {
		return nil, errData
	}
	result["items"] = items
	return result, nil
}

func (this *GenericRepository) ExecuteKeyValueList(keyAlias string, valueAlias string, query string, params ...any) ([]map[string]any, error) {
	var rows *sql.Rows
	var err error
	if len(params) == 0 {
		rows, err = this.db.Raw(query).Rows()
	} else {
		rows, err = this.db.Raw(query, params...).Rows()
	}
	//rows, err := this.db.Raw(query, params...).Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	// Prepare a slice to hold the results
	var result = make([]map[string]any, 0) //Empty slide
	columns, err := rows.Columns()
	if err != nil {
		return nil, err
	}
	if len(columns) < 2 {
		return nil, fmt.Errorf("query is required 2 selected columns")
	}
	for rows.Next() {
		// Create a slice of interface{} to hold each column value
		columnsData := make([]interface{}, len(columns))
		columnPointers := make([]interface{}, len(columns))
		for i := range columnsData {
			columnPointers[i] = &columnsData[i]
		}
		// Scan the row into the slice
		if ex := rows.Scan(columnPointers...); ex != nil {
			return nil, ex
		}
		// Create a map to hold column names and their values
		rowMap := make(map[string]interface{})
		// Convert []byte to string to avoid base64 encoding in JSON
		if b, ok := columnsData[0].([]byte); ok {
			rowMap[keyAlias] = string(b)
		} else {
			rowMap[keyAlias] = columnsData[0]
		}
		if b, ok := columnsData[1].([]byte); ok {
			rowMap[valueAlias] = string(b)
		} else {
			rowMap[valueAlias] = columnsData[1]
		}
		// Add the map to the result slice
		result = append(result, rowMap)
	}
	return result, nil
}

func (this *GenericRepository) ExecuteJsonObject(query string, params ...any) (map[string]any, error) {
	var rows *sql.Rows
	var err error
	if len(params) == 0 {
		rows, err = this.db.Raw(query).Rows()
	} else {
		rows, err = this.db.Raw(query, params...).Rows()
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	columns, err := rows.Columns()
	if err != nil {
		return nil, err
	}
	rowMap := make(map[string]interface{})
	if rows.Next() {
		// Create a slice of interface{} to hold each column value
		columnsData := make([]interface{}, len(columns))
		columnPointers := make([]interface{}, len(columns))
		for i := range columnsData {
			columnPointers[i] = &columnsData[i]
		}
		// Scan the row into the slice
		if ex := rows.Scan(columnPointers...); ex != nil {
			return nil, ex
		}
		for i, col := range columns {
			jsonField := fxstring.ToJsonCase(col)
			// Convert []byte to string to avoid base64 encoding in JSON
			if b, ok := columnsData[i].([]byte); ok {
				rowMap[jsonField] = string(b)
			} else {
				rowMap[jsonField] = columnsData[i]
			}
		}
	}
	return rowMap, nil
}

func (this *GenericRepository) ExecuteStringList(query string, params ...any) ([]string, error) {
	var rows *sql.Rows
	var err error
	if len(params) == 0 {
		rows, err = this.db.Raw(query).Rows()
	} else {
		rows, err = this.db.Raw(query, params...).Rows()
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	columns, err := rows.Columns()
	if err != nil {
		return nil, err
	}
	var result []string
	for rows.Next() {
		// Create a slice of interface{} to hold each column value
		columnsData := make([]interface{}, len(columns))
		columnPointers := make([]interface{}, len(columns))
		for i := range columnsData {
			columnPointers[i] = &columnsData[i]
		}
		// Scan the row into the slice
		if ex := rows.Scan(columnPointers...); ex != nil {
			return nil, ex
		}
		strValue := fxstring.ToString(columnsData[0])
		if strValue != "" {
			result = append(result, strValue)
		}
	}
	return result, nil
}

func (this *GenericRepository) ExecuteScalar(query string, params ...any) (any, error) {
	var rows *sql.Rows
	var err error
	if len(params) == 0 {
		rows, err = this.db.Raw(query).Rows()
	} else {
		rows, err = this.db.Raw(query, params...).Rows()
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	// Prepare a slice to hold the results
	var result interface{}
	columns, err := rows.Columns()
	if err != nil {
		return nil, err
	}
	if rows.Next() {
		// Create a slice of interface{} to hold each column value
		columnsData := make([]interface{}, len(columns))
		columnPointers := make([]interface{}, len(columns))
		for i := range columnsData {
			columnPointers[i] = &columnsData[i]
		}
		// Scan the row into the slice
		if ex := rows.Scan(columnPointers...); ex != nil {
			return nil, ex
		}
		// Return the first column value in the first row
		result = columnsData[0]
		return result, nil
	}
	return result, nil
}

func (this *GenericRepository) ExecuteScalarAsBool(query string, params ...any) (bool, error) {
	var value any
	var err error
	if len(params) == 0 {
		value, err = this.ExecuteScalar(query)
	} else {
		value, err = this.ExecuteScalar(query, params...)
	}
	if err != nil {
		return false, err
	}
	if value == nil {
		return false, nil
	}
	switch v := value.(type) {
	case bool:
		return v, nil
	case int:
		return v != 0, nil
	case int64:
		return v != 0, nil
	case float64:
		return v != 0, nil
	case []byte:
		return this.parseBoolFromString(string(v))
	case string:
		return this.parseBoolFromString(v)
	default:
		return false, fmt.Errorf("unsupported type %T for boolean conversion", value)
	}
}

func (this *GenericRepository) ExecuteScalarAsLong(query string, params ...any) (int64, error) {
	var value any
	var err error
	if len(params) == 0 {
		value, err = this.ExecuteScalar(query)
	} else {
		value, err = this.ExecuteScalar(query, params...)
	}
	if err != nil {
		return 0, err
	}
	if value == nil {
		return 0, nil
	}
	switch v := value.(type) {
	case int64:
		return v, nil
	case int:
		return int64(v), nil
	case float64:
		return int64(v), nil
	case []uint8:
		i, convErr := strconv.ParseInt(string(v), 10, 64)
		if convErr != nil {
			return 0, fmt.Errorf("cannot parse value: %v", convErr)
		}
		return i, nil
	default:
		return 0, fmt.Errorf("unsupported type: %T", value)
	}
}

func (this *GenericRepository) ExecuteScalarAsString(query string, params ...any) (string, error) {
	var value any
	var err error
	if len(params) == 0 {
		value, err = this.ExecuteScalar(query)
	} else {
		value, err = this.ExecuteScalar(query, params...)
	}
	if err != nil {
		return "", err
	}
	str := fmt.Sprintf("%v", value)
	return str, nil
}

func (this *GenericRepository) Create(model any) (any, error) {
	result := this.db.Create(model)
	if result.Error != nil {
		return nil, result.Error
	}
	return model, nil
}

func (this *GenericRepository) Save(record any) (any, error) {
	// Use db.Save for upsert behavior (Insert or Update if record already exists)
	result := this.db.Save(record)
	if result.Error != nil {
		return nil, result.Error
	}
	return record, nil
}

func (this *GenericRepository) Delete(model any, conditions ...any) (int64, error) {
	result := this.db.Delete(model, conditions...)
	if result.Error != nil {
		return 0, result.Error
	}
	return result.RowsAffected, nil
}

func (this *GenericRepository) DeleteAll(models []any) (int64, error) {
	var rowsEffected int64
	for _, model := range models {
		result := this.db.Delete(model)
		if result.Error != nil {
			return 0, fmt.Errorf("failed to delete model %v: %w", model, result.Error)
		}
		rowsEffected += result.RowsAffected
	}
	return rowsEffected, nil
}

// ----------------------------------------------------------------------------------------
// Generic (typed) methods — Go 1.27 generic methods. Consolidated from the former
// fxsql package. Not part of IGenericRepository (interfaces cannot declare type
// parameters); reach them via the concrete *GenericRepository.
// ----------------------------------------------------------------------------------------

// FindFirst returns the first row where fieldName == fieldValue, scanned into T.
func (this *GenericRepository) FindFirst[T any](fieldName string, fieldValue any) (T, error) {
	var result T
	params := map[string]any{fieldName: fieldValue}
	if err := this.db.Where(params).First(&result).Error; err != nil {
		return result, err
	}
	return result, nil
}

// FindOptionalObjectById looks up a row by primary key (uuid.UUID or its string
// form), returning an empty Optional when not found or when id is not a UUID.
func (this *GenericRepository) FindOptionalObjectById[T any](id any) (*fxmodel.Optional[T], error) {
	var result T
	var inputId uuid.UUID
	if reflect.TypeOf(id) == reflect.TypeOf(uuid.UUID{}) {
		inputId = id.(uuid.UUID)
	} else if s, ok := id.(string); ok {
		parsedUUID, err := uuid.Parse(s)
		if err != nil {
			return &fxmodel.Optional[T]{Value: nil}, nil
		}
		inputId = parsedUUID
	} else {
		return &fxmodel.Optional[T]{Value: nil}, nil
	}
	if err := this.db.First(&result, inputId).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return &fxmodel.Optional[T]{Value: nil}, nil
		}
		return &fxmodel.Optional[T]{Value: nil}, err
	}
	return &fxmodel.Optional[T]{Value: &result}, nil
}

// ExecuteModelList runs a raw query and scans all rows into []T.
func (this *GenericRepository) ExecuteModelList[T any](query string, params ...any) ([]T, error) {
	var results []T
	tx := this.db.Raw(query, params...).Scan(&results)
	if tx.Error != nil {
		if errors.Is(tx.Error, gorm.ErrRecordNotFound) {
			return results, nil
		}
		return nil, fmt.Errorf("error executing query: %w", tx.Error)
	}
	return results, nil
}

// ExecuteModelObject runs a raw query and scans the first row into T. A missing
// row returns the zero value of T and a nil error.
func (this *GenericRepository) ExecuteModelObject[T any](query string, params ...any) (T, error) {
	var result T
	if err := this.db.Raw(query, params...).First(&result).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return result, nil
		}
		return result, err
	}
	return result, nil
}

// ExecuteOptionalObject runs a raw query and returns the first row as an
// Optional[T], empty when no row matches.
func (this *GenericRepository) ExecuteOptionalObject[T any](query string, params ...any) (fxmodel.Optional[T], error) {
	var result T
	if err := this.db.Raw(query, params...).First(&result).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fxmodel.Optional[T]{Value: nil}, nil
		}
		return fxmodel.Optional[T]{Value: nil}, err
	}
	return fxmodel.Optional[T]{Value: &result}, nil
}

// DeleteModels deletes each model in the slice and returns the total rows affected.
func (this *GenericRepository) DeleteModels[T any](models []T) (int64, error) {
	var rowsEffected int64
	for _, model := range models {
		result := this.db.Delete(model)
		if result.Error != nil {
			return 0, fmt.Errorf("failed to delete model %v: %w", model, result.Error)
		}
		rowsEffected += result.RowsAffected
	}
	return rowsEffected, nil
}

// ----------------------------------------------------------------------------------------
// Private functions
// ----------------------------------------------------------------------------------------
func (this *GenericRepository) buildLimitingClause(page fxmodel.Pageable) string {
	offset := page.PageSize * (page.PageNumber - 1)
	result := fmt.Sprint(" LIMIT ", page.PageSize, " OFFSET ", offset)
	return result
}

func (this *GenericRepository) buildSortingClause(order string) string {
	if order == "" {
		return ""
	}
	var sortBuilder strings.Builder
	sortBuilder.WriteString(" ORDER BY ")
	orderBy := order
	pairs := strings.Fields(order)
	if len(pairs) == 1 {
		orderBy = orderBy + " asc"
	}
	sortBuilder.WriteString(orderBy)
	result := fmt.Sprintf(" %s ", sortBuilder.String())
	return result
}

func (this *GenericRepository) buildCountingQuery(query string) string {
	result := fmt.Sprintf("Select count(0) from (%s) alias", query)
	return result
}

func (this *GenericRepository) parseBoolFromString(s string) (bool, error) {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "1", "true", "yes":
		return true, nil
	case "0", "false", "no":
		return false, nil
	default:
		return false, fmt.Errorf("cannot convert string %q to bool", s)
	}
}
