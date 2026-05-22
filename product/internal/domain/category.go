package domain

import "time"

type Category struct {
	ID          string
	Name        string
	Description string
	CreatedAt   time.Time
}
