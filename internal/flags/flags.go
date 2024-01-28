package flags

import "time"

// Flag represents database table of Flags.
type Flag struct {
	ID        int       `db:"id"`
	UserUUID  string    `db:"user_uuid"`
	Name      string    `db:"name"`
	IsEnabled bool      `db:"is_enabled"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}
