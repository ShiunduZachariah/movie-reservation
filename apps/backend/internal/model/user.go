package model

type UserRole string

const (
	RoleUser  UserRole = "user"
	RoleAdmin UserRole = "admin"
)

type User struct {
	Base
	ClerkID string   `db:"clerk_id" json:"clerk_id"`
	Email   string   `db:"email" json:"email"`
	Name    string   `db:"name" json:"name"`
	Role    UserRole `db:"role" json:"role"`
}
