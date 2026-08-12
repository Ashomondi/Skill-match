package models 

import "time"

type User struct {
	ID  string  `json""id"`
	Email  string  `json:"email"`
	Password  string  `json:"-"`
	CreatedAt  time.time  `json:"created_at"`
	UpdatedAt  time.time `json:"updated_at"`
}