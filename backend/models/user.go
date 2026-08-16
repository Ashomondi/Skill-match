package models

import "time"

// User is the canonical domain representation of a registered user. The
// Password field holds the bcrypt hash in persistence and is never
// serialized to clients (json:"-").
type User struct {
	ID        string    `json:"id"`
	Email     string    `json:"email"`
	Password  string    `json:"-"`
	FullName  string    `json:"fullName"`
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
