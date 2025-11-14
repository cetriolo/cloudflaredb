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

// scanUser scans a user row from database results using position-based scanning.
// Handles type differences between SQLite3 (int64) and D1 driver (float64) by scanning
// integers into float64 and converting to int64.
//
// Expected column order: id, external_id, created_at, updated_at
func scanUser(scanner interface {
	Scan(dest ...interface{}) error
}, id *int64, externalID *string, createdAt, updatedAt *time.Time) error {
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

// scanRoom scans a room row from database results using position-based scanning.
// Handles type differences between SQLite3 (int64) and D1 driver (float64) by scanning
// integers into float64 and converting to int64.
//
// Expected column order: id, name, description, room_type_id, created_at, updated_at
func scanRoom(scanner interface {
	Scan(dest ...interface{}) error
}, id *int64, name, description *string, roomTypeID **int64, createdAt, updatedAt *time.Time) error {
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
		} else if i, ok := roomTypeIDFloat.(int64); ok {
			*roomTypeID = &i
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

// scanRoomType scans a room type row from database results using position-based scanning.
// Handles type differences between SQLite3 (int64) and D1 driver (float64) by scanning
// integers into float64 and converting to int64.
//
// Expected column order: id, size, style, created_at, updated_at
func scanRoomType(scanner interface {
	Scan(dest ...interface{}) error
}, id *int64, size, style *string, createdAt, updatedAt *time.Time) error {
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
