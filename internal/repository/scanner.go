package repository

import (
	"database/sql/driver"
	"fmt"
	"time"
)

// NullableTime is a custom type that can scan both time.Time and string timestamps
// This is needed because the cfd1 driver returns timestamps as strings
type NullableTime struct {
	Time  time.Time
	Valid bool
}

// Scan implements the sql.Scanner interface
func (nt *NullableTime) Scan(value interface{}) error {
	if value == nil {
		nt.Valid = false
		return nil
	}

	nt.Valid = true

	switch v := value.(type) {
	case time.Time:
		nt.Time = v
		return nil
	case string:
		// Try parsing common timestamp formats
		formats := []string{
			time.RFC3339,
			time.RFC3339Nano,
			"2006-01-02 15:04:05",
			"2006-01-02T15:04:05",
			"2006-01-02 15:04:05.999999999",
			"2006-01-02T15:04:05.999999999",
		}

		var err error
		for _, format := range formats {
			nt.Time, err = time.Parse(format, v)
			if err == nil {
				return nil
			}
		}
		return fmt.Errorf("failed to parse timestamp string: %s", v)
	case []byte:
		// Try parsing as string
		return nt.Scan(string(v))
	default:
		return fmt.Errorf("unsupported type for NullableTime: %T", value)
	}
}

// Value implements the driver.Valuer interface
func (nt NullableTime) Value() (driver.Value, error) {
	if !nt.Valid {
		return nil, nil
	}
	return nt.Time, nil
}

// ColumnScanner is an interface that can return column names
type ColumnScanner interface {
	Scan(dest ...interface{}) error
	Columns() ([]string, error)
}

// scanUser is a helper function to scan a user row that handles cfd1 timestamp strings, numeric types, and column ordering bugs
func scanUser(scanner interface {
	Scan(dest ...interface{}) error
}, id *int64, email, name *string, createdAt, updatedAt *time.Time) error {
	// Try to get column names if available (for *sql.Rows)
	if colScanner, ok := scanner.(ColumnScanner); ok {
		return scanUserWithColumns(colScanner, id, email, name, createdAt, updatedAt)
	}

	// Fallback to position-based scanning for *sql.Row
	var createdAtNT, updatedAtNT NullableTime
	var idFloat float64

	err := scanner.Scan(&idFloat, email, name, &createdAtNT, &updatedAtNT)
	if err != nil {
		return err
	}

	*id = int64(idFloat)

	if createdAtNT.Valid {
		*createdAt = createdAtNT.Time
	}
	if updatedAtNT.Valid {
		*updatedAt = updatedAtNT.Time
	}

	return nil
}

// scanUserWithColumns scans a user using column names to handle cfd1's column ordering bug
func scanUserWithColumns(scanner ColumnScanner, id *int64, email, name *string, createdAt, updatedAt *time.Time) error {
	cols, err := scanner.Columns()
	if err != nil {
		return err
	}

	// Create a map to store values by column name
	values := make([]interface{}, len(cols))
	for i := range values {
		values[i] = new(interface{})
	}

	err = scanner.Scan(values...)
	if err != nil {
		return err
	}

	// Map values to struct fields based on column names
	for i, col := range cols {
		val := *(values[i].(*interface{}))
		if val == nil {
			continue
		}

		switch col {
		case "id":
			if f, ok := val.(float64); ok {
				*id = int64(f)
			}
		case "email":
			if s, ok := val.(string); ok {
				*email = s
			}
		case "name":
			if s, ok := val.(string); ok {
				*name = s
			}
		case "created_at":
			*createdAt = parseTimeValue(val)
		case "updated_at":
			*updatedAt = parseTimeValue(val)
		}
	}

	return nil
}

// scanRoom is a helper function to scan a room row that handles cfd1 timestamp strings, numeric types, and column ordering bugs
func scanRoom(scanner interface {
	Scan(dest ...interface{}) error
}, id *int64, name, description *string, capacity *int, createdAt, updatedAt *time.Time) error {
	// Try to get column names if available (for *sql.Rows)
	if colScanner, ok := scanner.(ColumnScanner); ok {
		return scanRoomWithColumns(colScanner, id, name, description, capacity, createdAt, updatedAt)
	}

	// Fallback to position-based scanning for *sql.Row
	var createdAtNT, updatedAtNT NullableTime
	var idFloat, capacityFloat float64

	err := scanner.Scan(&idFloat, name, description, &capacityFloat, &createdAtNT, &updatedAtNT)
	if err != nil {
		return err
	}

	*id = int64(idFloat)
	*capacity = int(capacityFloat)

	if createdAtNT.Valid {
		*createdAt = createdAtNT.Time
	}
	if updatedAtNT.Valid {
		*updatedAt = updatedAtNT.Time
	}

	return nil
}

// scanRoomWithColumns scans a room using column names to handle cfd1's column ordering bug
func scanRoomWithColumns(scanner ColumnScanner, id *int64, name, description *string, capacity *int, createdAt, updatedAt *time.Time) error {
	cols, err := scanner.Columns()
	if err != nil {
		return err
	}

	// Create a map to store values by column name
	values := make([]interface{}, len(cols))
	for i := range values {
		values[i] = new(interface{})
	}

	err = scanner.Scan(values...)
	if err != nil {
		return err
	}

	// Map values to struct fields based on column names
	for i, col := range cols {
		val := *(values[i].(*interface{}))
		if val == nil {
			continue
		}

		switch col {
		case "id":
			if f, ok := val.(float64); ok {
				*id = int64(f)
			}
		case "name":
			if s, ok := val.(string); ok {
				*name = s
			}
		case "description":
			if s, ok := val.(string); ok {
				*description = s
			}
		case "capacity":
			if f, ok := val.(float64); ok {
				*capacity = int(f)
			}
		case "created_at":
			*createdAt = parseTimeValue(val)
		case "updated_at":
			*updatedAt = parseTimeValue(val)
		}
	}

	return nil
}

// parseTimeValue parses a time value from various types
func parseTimeValue(val interface{}) time.Time {
	switch v := val.(type) {
	case time.Time:
		return v
	case string:
		// Try parsing common timestamp formats
		formats := []string{
			time.RFC3339,
			time.RFC3339Nano,
			"2006-01-02 15:04:05",
			"2006-01-02T15:04:05",
			"2006-01-02 15:04:05.999999999",
			"2006-01-02T15:04:05.999999999",
		}

		for _, format := range formats {
			if t, err := time.Parse(format, v); err == nil {
				return t
			}
		}
	}
	return time.Time{}
}
