package models

import "time"

type User struct {
	ID        int64
	Name      string
	Email     string
	CreatedAt time.Time
	UpdatedAt time.Time
}

type UserWithPassword struct {
	User
	PasswordHash string
}
