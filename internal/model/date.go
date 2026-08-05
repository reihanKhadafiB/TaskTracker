package model

import (
	"database/sql/driver"
	"fmt"
	"time"
)

type Date struct {
	time.Time
}

const dateLayout = "2006-01-02"

func (d Date) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`"%s"`, d.Time.Format(dateLayout))), nil
}

func (d *Date) UnmarshalJSON(data []byte) error {
	s := string(data)
	if s == "null" {
		return nil
	}
	if len(s) < 2 || s[0] != '"' || s[len(s)-1] != '"' {
		return fmt.Errorf("date must be a JSON string, got %s", s)
	}
	s = s[1 : len(s)-1]
	if s == "" {
		return nil
	}
	t, err := time.Parse(dateLayout, s)
	if err != nil {
		return fmt.Errorf("invalid date format, expected YYYY-MM-DD: %w", err)
	}
	d.Time = t
	return nil
}

func (d *Date) Scan(value any) error {
	if value == nil {
		return nil
	}
	t, ok := value.(time.Time)
	if !ok {
		return fmt.Errorf("cannot scan type %T into model.Date", value)
	}
	d.Time = t
	return nil
}

func (d Date) Value() (driver.Value, error) {
	return d.Time, nil
}
