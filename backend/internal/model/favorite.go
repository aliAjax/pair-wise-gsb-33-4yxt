package model

import "time"

// Favorite links a user to a plant or an article.
// GroupID nil means the favorite stays in the virtual "未分类" group.
type Favorite struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	UserID     uint      `gorm:"index:idx_fav_user_target,unique;not null" json:"user_id"`
	TargetType string    `gorm:"size:16;index:idx_fav_user_target,unique;not null" json:"target_type"`
	TargetID   uint      `gorm:"index:idx_fav_user_target,unique;not null" json:"target_id"`
	GroupID    *uint     `gorm:"index:idx_fav_group;default:null" json:"group_id"`
	CreatedAt  time.Time `json:"created_at"`
}
