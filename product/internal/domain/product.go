package domain

import "time"

type Product struct {
	ID           string
	Name         string
	Description  string
	CategoryID   string
	Price        float64
	Stock        int
	Manufacturer string
	RequiresRX   bool
	ImageURL     string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}
