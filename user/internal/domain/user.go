package domain

import "time"

type User struct {
	ID           string
	Email        string
	PasswordHash string
	FirstName    string
	LastName     string
	Role         string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type NotificationLog struct {
	ID      string
	UserID  string
	Type    string
	Subject string
	Body    string
	SentAt  time.Time
	Status  string
}
