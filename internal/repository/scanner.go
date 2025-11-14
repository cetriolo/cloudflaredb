package repository

import (
	"database/sql/driver"
	"fmt"
	"time"
)

// NullableTime is a custom type that handles timestamp scanning across different SQL drivers.
// This type is essential for compatibility with both SQLite and Cloudflare D1:
//   - SQLite returns timestamps as time.Time objects
//   - Cloudflare D1 (cfd1 driver) returns timestamps as strings
//
// It implements sql.Scanner and driver.Valuer for seamless database integration.
type NullableTime struct {
	Time  time.Time // The actual time value
	Valid bool      // Indicates whether the time value is valid (not NULL)
}

// Scan implements the sql.Scanner interface for NullableTime.
// It handles multiple input types:
//   - nil: sets Valid to false
//   - time.Time: directly assigns the value
//   - string: parses using multiple timestamp formats
//   - []byte: converts to string and parses
//
// Returns an error if the value cannot be parsed or is an unsupported type.
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

// Value implements the driver.Valuer interface for NullableTime.
// Returns nil if the time is not valid, otherwise returns the time.Time value.
// This allows NullableTime to be used in SQL INSERT and UPDATE statements.
func (nt NullableTime) Value() (driver.Value, error) {
	if !nt.Valid {
		return nil, nil
	}
	return nt.Time, nil
}

// ColumnScanner is an interface that extends the standard sql.Scanner with column name access.
// This is typically implemented by *sql.Rows but not by *sql.Row.
// Column name access is crucial for handling Cloudflare D1's column ordering inconsistencies.
type ColumnScanner interface {
	Scan(dest ...interface{}) error // Scans row data into destinations
	Columns() ([]string, error)     // Returns the column names in order
}

// scanUser is a helper function to scan a user row from database results.
// It handles multiple database driver quirks:
//   - Cloudflare D1 returns timestamps as strings (handled via NullableTime)
//   - Cloudflare D1 returns integers as float64 (requires type conversion)
//   - Cloudflare D1 may return columns in unexpected order (uses column names when available)
//
// The function attempts to use column-name-based scanning when available (sql.Rows),
// and falls back to position-based scanning for sql.Row.
func scanUser(scanner interface {
	Scan(dest ...interface{}) error
}, id *int64, externalID *string, createdAt, updatedAt *time.Time) error {
	// Try to get column names if available (for *sql.Rows).
	// This approach is more robust against column ordering differences between databases.
	if colScanner, ok := scanner.(ColumnScanner); ok {
		return scanUserWithColumns(colScanner, id, externalID, createdAt, updatedAt)
	}

	// Fallback to position-based scanning for *sql.Row.
	// Assumes columns are in the order: id, external_id, created_at, updated_at
	var createdAtNT, updatedAtNT NullableTime
	var idFloat float64

	err := scanner.Scan(&idFloat, externalID, &createdAtNT, &updatedAtNT)
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

// scanUserWithColumns scans a user using column names for accurate field mapping.
// This approach is essential when working with Cloudflare D1, which may return columns
// in a different order than expected. By mapping values using column names instead of
// positions, we ensure correctness regardless of column order.
//
// The function:
//  1. Retrieves column names from the scanner
//  2. Scans all values into interface{} pointers
//  3. Maps each value to the appropriate struct field based on column name
//  4. Handles type conversions (float64 to int64 for IDs, string/time.Time for timestamps)
func scanUserWithColumns(scanner ColumnScanner, id *int64, externalID *string, createdAt, updatedAt *time.Time) error {
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
			switch v := val.(type) {
			case float64:
				*id = int64(v)
			case int64:
				*id = v
			case int:
				*id = int64(v)
			}
		case "external_id":
			if s, ok := val.(string); ok {
				*externalID = s
			}
		case "created_at":
			*createdAt = parseTimeValue(val)
		case "updated_at":
			*updatedAt = parseTimeValue(val)
		}
	}

	return nil
}

// scanRoom is a helper function to scan a room row from database results.
// It handles the same database driver quirks as scanUser:
//   - Cloudflare D1 timestamp strings (via NullableTime)
//   - Cloudflare D1 numeric type conversions (float64 to int64)
//   - Cloudflare D1 column ordering issues (uses column names when available)
//
// The function attempts to use column-name-based scanning when available (sql.Rows),
// and falls back to position-based scanning for sql.Row.
func scanRoom(scanner interface {
	Scan(dest ...interface{}) error
}, id *int64, name, description *string, roomTypeID **int64, createdAt, updatedAt *time.Time) error {
	// Try to get column names if available (for *sql.Rows).
	// This approach is more robust against column ordering differences between databases.
	if colScanner, ok := scanner.(ColumnScanner); ok {
		return scanRoomWithColumns(colScanner, id, name, description, roomTypeID, createdAt, updatedAt)
	}

	// Fallback to position-based scanning for *sql.Row.
	// Assumes columns are in the order: id, name, description, room_type_id, created_at, updated_at
	var createdAtNT, updatedAtNT NullableTime
	var idFloat float64
	var roomTypeIDFloat interface{}

	err := scanner.Scan(&idFloat, name, description, &roomTypeIDFloat, &createdAtNT, &updatedAtNT)
	if err != nil {
		return err
	}

	*id = int64(idFloat)

	// Handle nullable room_type_id
	if roomTypeIDFloat != nil {
		if f, ok := roomTypeIDFloat.(float64); ok {
			val := int64(f)
			*roomTypeID = &val
		}
	}

	if createdAtNT.Valid {
		*createdAt = createdAtNT.Time
	}
	if updatedAtNT.Valid {
		*updatedAt = updatedAtNT.Time
	}

	return nil
}

// scanRoomWithColumns scans a room using column names for accurate field mapping.
// This approach is essential when working with Cloudflare D1, which may return columns
// in a different order than expected. By mapping values using column names instead of
// positions, we ensure correctness regardless of column order.
//
// The function:
//  1. Retrieves column names from the scanner
//  2. Scans all values into interface{} pointers
//  3. Maps each value to the appropriate struct field based on column name
//  4. Handles type conversions (float64 to int64, string/time.Time for timestamps)
func scanRoomWithColumns(scanner ColumnScanner, id *int64, name, description *string, roomTypeID **int64, createdAt, updatedAt *time.Time) error {
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
			switch v := val.(type) {
			case float64:
				*id = int64(v)
			case int64:
				*id = v
			case int:
				*id = int64(v)
			}
		case "name":
			if s, ok := val.(string); ok {
				*name = s
			}
		case "description":
			if s, ok := val.(string); ok {
				*description = s
			}
		case "room_type_id":
			switch v := val.(type) {
			case float64:
				val := int64(v)
				*roomTypeID = &val
			case int64:
				*roomTypeID = &v
			case int:
				val := int64(v)
				*roomTypeID = &val
			}
		case "created_at":
			*createdAt = parseTimeValue(val)
		case "updated_at":
			*updatedAt = parseTimeValue(val)
		}
	}

	return nil
}

// parseTimeValue parses a time value from various types.
// It handles both native time.Time values and string timestamps.
// Supports multiple timestamp formats commonly used in SQL databases:
//   - RFC3339 and RFC3339Nano (standard Go time formats)
//   - SQL standard formats: "2006-01-02 15:04:05"
//   - ISO 8601 formats: "2006-01-02T15:04:05"
//   - Formats with fractional seconds
//
// Returns a zero time.Time{} if parsing fails for all formats.
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

// scanRoomType is a helper function to scan a room type row from database results.
// It handles the same database driver quirks as scanUser and scanRoom:
//   - Cloudflare D1 timestamp strings (via NullableTime)
//   - Cloudflare D1 numeric type conversions (float64 to int64)
//   - Cloudflare D1 column ordering issues (uses column names when available)
//
// The function attempts to use column-name-based scanning when available (sql.Rows),
// and falls back to position-based scanning for sql.Row.
func scanRoomType(scanner interface {
	Scan(dest ...interface{}) error
}, id *int64, size, style *string, createdAt, updatedAt *time.Time) error {
	// Try to get column names if available (for *sql.Rows).
	// This approach is more robust against column ordering differences between databases.
	if colScanner, ok := scanner.(ColumnScanner); ok {
		return scanRoomTypeWithColumns(colScanner, id, size, style, createdAt, updatedAt)
	}

	// Fallback to position-based scanning for *sql.Row.
	// Assumes columns are in the order: id, size, style, created_at, updated_at
	var createdAtNT, updatedAtNT NullableTime
	var idFloat float64

	err := scanner.Scan(&idFloat, size, style, &createdAtNT, &updatedAtNT)
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

// scanRoomTypeWithColumns scans a room type using column names for accurate field mapping.
// This approach is essential when working with Cloudflare D1, which may return columns
// in a different order than expected. By mapping values using column names instead of
// positions, we ensure correctness regardless of column order.
//
// The function:
//  1. Retrieves column names from the scanner
//  2. Scans all values into interface{} pointers
//  3. Maps each value to the appropriate struct field based on column name
//  4. Handles type conversions (float64 to int64 for IDs, string/time.Time for timestamps)
func scanRoomTypeWithColumns(scanner ColumnScanner, id *int64, size, style *string, createdAt, updatedAt *time.Time) error {
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
			switch v := val.(type) {
			case float64:
				*id = int64(v)
			case int64:
				*id = v
			case int:
				*id = int64(v)
			}
		case "size":
			if s, ok := val.(string); ok {
				*size = s
			}
		case "style":
			if s, ok := val.(string); ok {
				*style = s
			}
		case "created_at":
			*createdAt = parseTimeValue(val)
		case "updated_at":
			*updatedAt = parseTimeValue(val)
		}
	}

	return nil
}
