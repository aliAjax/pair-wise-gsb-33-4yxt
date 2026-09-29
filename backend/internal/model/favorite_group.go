package model

import "time"

// FavoriteGroup is a user-owned folder that favorites can be assigned to.
// Favorites without a group belong to the virtual "未分类" group (group_id NULL).
type FavoriteGroup struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	UserID    uint      `gorm:"index:idx_fav_group_user_name,unique;not null" json:"user_id"`
	Name      string    `gorm:"size:64;index:idx_fav_group_user_name,unique;not null" json:"name"`
	CreatedAt time.Time `json:"created_at"`
}
