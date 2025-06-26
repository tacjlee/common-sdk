package fxrepositories

import (
	"fmt"
	"github.com/tacjlee/common-sdk/packages/fxmodels"
	"github.com/tacjlee/common-sdk/packages/fxstrings"
	"gorm.io/gorm"
	"math"
	"strings"
)

type IGenericRepository interface {
	GetDB() *gorm.DB
	ExecuteNonQuery(command string, params ...any) (int64, error)
	ExecuteJsonList(query string, params ...any) ([]map[string]any, error)
	ExecuteJsonPaging(query string, pageable fxmodels.Pageable, params ...any) (map[string]any, error)
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
	result := this.db.Exec(query, params...)
	if result.Error != nil {
		return 0, result.Error
	}
	rowsAffected := result.RowsAffected
	return rowsAffected, nil
}

func (this *genericRepository) ExecuteJsonList(query string, params ...any) ([]map[string]any, error) {
	rows, err := this.db.Raw(query, params...).Rows()
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
			jsonField := fxstrings.ToJsonCase(col)
			rowMap[jsonField] = columnsData[i]
		}
		// Add the map to the result slice
		result = append(result, rowMap)
	}
	return result, nil
}

func (this *genericRepository) ExecuteJsonPaging(query string, pageable fxmodels.Pageable, params ...any) (map[string]any, error) {
	countingSql := this.buildCountingQuery(query)
	totalItems, err := this.ExecuteScalarAsLong(countingSql, params...)
	if err != nil {
		return nil, err
	}
	totalPages := int(math.Ceil(float64(totalItems) / float64(pageable.PageSize)))
	result := make(map[string]any)
	result["totalItems"] = totalItems
	result["totalPages"] = totalPages
	result["pageSize"] = pageable.PageSize
	result["currentPage"] = pageable.PageNumber
	result["items"] = make([]map[string]any, 0)

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
	rows, err := this.db.Raw(query, params...).Rows()
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
		rowMap[keyAlias] = columnsData[0]
		rowMap[valueAlias] = columnsData[1]
		// Add the map to the result slice
		result = append(result, rowMap)
	}
	return result, nil
}

func (this *genericRepository) ExecuteJsonObject(query string, params ...any) (map[string]any, error) {
	rows, err := this.db.Raw(query, params...).Rows()
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
			jsonField := fxstrings.ToJsonCase(col)
			rowMap[jsonField] = columnsData[i]
		}
	}
	return rowMap, nil
}

func (this *genericRepository) ExecuteStringList(query string, params ...any) ([]string, error) {
	rows, err := this.db.Raw(query, params...).Rows()
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
		strValue := fxstrings.ToString(columnsData[0])
		if strValue != "" {
			result = append(result, strValue)
		}
	}
	return result, nil
}

func (this *genericRepository) ExecuteScalar(query string, params ...any) (any, error) {
	rows, err := this.db.Raw(query, params...).Rows()
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
	value, err := this.ExecuteScalar(query, params...)
	if err != nil {
		return false, err
	}
	str := fmt.Sprintf("%v", value)
	if str == "1" || str == "true" {
		return true, nil
	} else {
		return false, nil
	}
}

func (this *genericRepository) ExecuteScalarAsLong(query string, params ...any) (int64, error) {
	value, err := this.ExecuteScalar(query, params...)
	if err != nil {
		return 0, err
	}
	return value.(int64), nil
}

func (this *genericRepository) ExecuteScalarAsString(query string, params ...any) (string, error) {
	value, err := this.ExecuteScalar(query, params...)
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
func (this *genericRepository) buildLimitingClause(page fxmodels.Pageable) string {
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
