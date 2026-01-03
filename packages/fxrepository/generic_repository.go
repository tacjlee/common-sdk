package fxrepository

import (
	"database/sql"
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/tacjlee/common-sdk/packages/fxmodel"
	"github.com/tacjlee/common-sdk/packages/fxstring"
	"gorm.io/gorm"
)

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
type genericRepository struct {
	db *gorm.DB
}

func NewGenericRepository(db *gorm.DB) IGenericRepository {
	return &genericRepository{db: db}
}

func (this *genericRepository) GetDB() *gorm.DB {
	return this.db
}

// For update, delete command
func (this *genericRepository) ExecuteNonQuery(query string, params ...any) (int64, error) {
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

func (this *genericRepository) ExecuteJsonList(query string, params ...any) ([]map[string]any, error) {
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

func (this *genericRepository) ExecuteJsonPaging(query string, pageable fxmodel.Pageable, params ...any) (map[string]any, error) {
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

func (this *genericRepository) ExecuteKeyValueList(keyAlias string, valueAlias string, query string, params ...any) ([]map[string]any, error) {
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

func (this *genericRepository) ExecuteJsonObject(query string, params ...any) (map[string]any, error) {
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

func (this *genericRepository) ExecuteStringList(query string, params ...any) ([]string, error) {
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

func (this *genericRepository) ExecuteScalar(query string, params ...any) (any, error) {
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

func (this *genericRepository) ExecuteScalarAsBool(query string, params ...any) (bool, error) {
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

func (this *genericRepository) ExecuteScalarAsLong(query string, params ...any) (int64, error) {
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

func (this *genericRepository) ExecuteScalarAsString(query string, params ...any) (string, error) {
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

func (this *genericRepository) Create(model any) (any, error) {
	result := this.db.Create(model)
	if result.Error != nil {
		return nil, result.Error
	}
	return model, nil
}

func (this *genericRepository) Save(record any) (any, error) {
	// Use db.Save for upsert behavior (Insert or Update if record already exists)
	result := this.db.Save(record)
	if result.Error != nil {
		return nil, result.Error
	}
	return record, nil
}

func (this *genericRepository) Delete(model any, conditions ...any) (int64, error) {
	result := this.db.Delete(model, conditions...)
	if result.Error != nil {
		return 0, result.Error
	}
	return result.RowsAffected, nil
}

func (this *genericRepository) DeleteAll(models []any) (int64, error) {
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
func (this *genericRepository) buildLimitingClause(page fxmodel.Pageable) string {
	offset := page.PageSize * (page.PageNumber - 1)
	result := fmt.Sprint(" LIMIT ", page.PageSize, " OFFSET ", offset)
	return result
}

func (this *genericRepository) buildSortingClause(order string) string {
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

func (this *genericRepository) buildCountingQuery(query string) string {
	result := fmt.Sprintf("Select count(0) from (%s) alias", query)
	return result
}

func (this *genericRepository) parseBoolFromString(s string) (bool, error) {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "1", "true", "yes":
		return true, nil
	case "0", "false", "no":
		return false, nil
	default:
		return false, fmt.Errorf("cannot convert string %q to bool", s)
	}
}
