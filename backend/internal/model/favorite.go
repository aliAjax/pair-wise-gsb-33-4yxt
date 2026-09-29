package model

import "time"

// Favorite links a user to a plant or an article.
// GroupID is nil when the favorite is uncategorized; it never carries a
// database foreign key so removing a group simply nulls it out and the
// favorited target keeps opening normally.
type Favorite struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	UserID     uint      `gorm:"index:idx_fav_user_target,unique;not null" json:"user_id"`
	TargetType string    `gorm:"size:16;index:idx_fav_user_target,unique;not null" json:"target_type"`
	TargetID   uint      `gorm:"index:idx_fav_user_target,unique;not null" json:"target_id"`
	GroupID    *uint     `gorm:"index;default:null" json:"group_id"`
	CreatedAt  time.Time `json:"created_at"`
}
