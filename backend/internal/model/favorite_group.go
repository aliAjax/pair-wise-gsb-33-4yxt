package model

import "time"

// FavoriteGroup is a user-owned folder for organizing favorites.
// Group names are unique per user; uncategorized favorites have no group.
type FavoriteGroup struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	UserID    uint      `gorm:"index:idx_fav_group_user_name,unique;not null" json:"user_id"`
	Name      string    `gorm:"size:64;index:idx_fav_group_user_name,unique;not null" json:"name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
