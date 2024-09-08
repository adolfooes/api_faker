package crud

import (
	"fmt"
	"strings"

	"github.com/adolfooes/api_faker/internal/db"
)

// Create inserts a new record into the table
func Create(table string, columns []string, values []interface{}) error {
	if len(columns) != len(values) {
		return fmt.Errorf("number of columns does not match the number of values")
	}

	columnsStr := strings.Join(columns, ", ")
	placeholders := make([]string, len(values))
	for i := range placeholders {
		placeholders[i] = fmt.Sprintf("$%d", i+1)
	}
	placeholdersStr := strings.Join(placeholders, ", ")

	query := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)", table, columnsStr, placeholdersStr)
	_, err := db.GetDB().Exec(query, values...)
	if err != nil {
		return fmt.Errorf("error creating record: %v", err)
	}
	return nil
}

// Read retrieves a record from the table based on the ID
func Read(table string, id int) (map[string]interface{}, error) {
	query := fmt.Sprintf("SELECT * FROM %s WHERE id = $1", table)

	// Use Query instead of QueryRow to get *sql.Rows
	rows, err := db.GetDB().Query(query, id)
	if err != nil {
		return nil, fmt.Errorf("error executing query: %v", err)
	}
	defer rows.Close()

	// Check if there are results
	if !rows.Next() {
		return nil, fmt.Errorf("no record found with id %d", id)
	}

	// Get column names dynamically
	columns, err := rows.Columns()
	if err != nil {
		return nil, fmt.Errorf("error getting columns: %v", err)
	}

	// Create slices to store values
	values := make([]interface{}, len(columns))
	valuePtrs := make([]interface{}, len(columns))
	for i := range values {
		valuePtrs[i] = &values[i]
	}

	// Scan the values into the slice of pointers
	err = rows.Scan(valuePtrs...)
	if err != nil {
		return nil, fmt.Errorf("error scanning row: %v", err)
	}

	// Create the map to store the result
	result := make(map[string]interface{})
	for i, col := range columns {
		result[col] = values[i]
	}

	return result, nil
}

// List retrieves records based on a table and a map of key and values
func List(table string, filters map[string]interface{}) ([]map[string]interface{}, error) {
	var whereClauses []string
	var args []interface{}
	i := 1
	for key, value := range filters {
		whereClauses = append(whereClauses, fmt.Sprintf("%s = $%d", key, i))
		args = append(args, value)
		i++
	}
	whereClause := strings.Join(whereClauses, " AND ")

	query := fmt.Sprintf("SELECT * FROM %s WHERE %s", table, whereClause)
	rows, err := db.GetDB().Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("error listing records: %v", err)
	}
	defer rows.Close()

	var results []map[string]interface{}
	columns, err := rows.Columns()
	if err != nil {
		return nil, fmt.Errorf("error getting columns: %v", err)
	}

	for rows.Next() {
		values := make([]interface{}, len(columns))
		valuePtrs := make([]interface{}, len(columns))
		for i := range columns {
			valuePtrs[i] = &values[i]
		}

		err := rows.Scan(valuePtrs...)
		if err != nil {
			return nil, fmt.Errorf("error reading record: %v", err)
		}

		result := make(map[string]interface{})
		for i, col := range columns {
			result[col] = values[i]
		}
		results = append(results, result)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over records: %v", err)
	}

	return results, nil
}

// Update updates a record based on a table and ID
func Update(table string, id int, updates map[string]interface{}) error {
	var setClauses []string
	var args []interface{}
	i := 1
	for key, value := range updates {
		setClauses = append(setClauses, fmt.Sprintf("%s = $%d", key, i))
		args = append(args, value)
		i++
	}
	setClause := strings.Join(setClauses, ", ")

	query := fmt.Sprintf("UPDATE %s SET %s WHERE id = $%d", table, setClause, i)
	args = append(args, id)
	_, err := db.GetDB().Exec(query, args...)
	if err != nil {
		return fmt.Errorf("error updating record: %v", err)
	}
	return nil
}

// Delete removes a record based on a table and ID
func Delete(table string, id int) error {
	query := fmt.Sprintf("DELETE FROM %s WHERE id = $1", table)
	_, err := db.GetDB().Exec(query, id)
	if err != nil {
		return fmt.Errorf("error deleting record: %v", err)
	}
	return nil
}

// Raw executes any SQL command (SELECT, INSERT, UPDATE, DELETE, etc.)
func Raw(query string, args ...interface{}) ([]map[string]interface{}, int64, error) {
	// Check if the query is a SELECT statement
	if isSelect(query) {
		// For SELECT queries, we return the results
		rows, err := db.GetDB().Query(query, args...)
		if err != nil {
			return nil, 0, fmt.Errorf("error executing query: %v", err)
		}
		defer rows.Close()

		// Get column names dynamically
		columns, err := rows.Columns()
		if err != nil {
			return nil, 0, fmt.Errorf("error getting columns: %v", err)
		}

		// Prepare a slice to store the results
		var results []map[string]interface{}

		// Iterate over the returned rows
		for rows.Next() {
			// Create slices to store the row's values
			values := make([]interface{}, len(columns))
			valuePtrs := make([]interface{}, len(columns))
			for i := range values {
				valuePtrs[i] = &values[i]
			}

			// Scan the row's values
			err := rows.Scan(valuePtrs...)
			if err != nil {
				return nil, 0, fmt.Errorf("error scanning row: %v", err)
			}

			// Store results in a map
			result := make(map[string]interface{})
			for i, col := range columns {
				result[col] = values[i]
			}

			results = append(results, result)
		}

		// Return the results from the SELECT query
		return results, 0, nil
	} else {
		// For INSERT, UPDATE, and DELETE, we use Exec, which does not return rows
		res, err := db.GetDB().Exec(query, args...)
		if err != nil {
			return nil, 0, fmt.Errorf("error executing non-select query: %v", err)
		}

		// Get the number of affected rows
		rowsAffected, err := res.RowsAffected()
		if err != nil {
			return nil, 0, fmt.Errorf("error getting rows affected: %v", err)
		}

		// For non-SELECT queries, we return nil in results and the number of affected rows
		return nil, rowsAffected, nil
	}
}

// Helper function to determine if the query is a SELECT statement
func isSelect(query string) bool {
	// Here we check if the query starts with "SELECT"
	// Adjust as needed for other SELECT variants.
	return len(query) >= 6 && query[:6] == "SELECT"
}
