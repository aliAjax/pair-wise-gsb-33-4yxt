package dto

// FavoriteRequest is the payload for adding a favorite.
// GroupID 0 (or omitted) places the favorite in uncategorized.
type FavoriteRequest struct {
	TargetType string `json:"target_type" binding:"required"`
	TargetID   uint   `json:"target_id" binding:"required"`
	GroupID    uint   `json:"group_id"`
}

// FavoriteGroupRequest is the payload for creating or renaming a group.
type FavoriteGroupRequest struct {
	Name string `json:"name" binding:"required,max=64"`
}

// FavoriteMoveRequest moves a favorite into or out of a group.
// GroupID 0 (or omitted) means uncategorized.
type FavoriteMoveRequest struct {
	GroupID uint `json:"group_id"`
}

// FavoriteGroupItem is a group with the real number of favorites in it.
type FavoriteGroupItem struct {
	ID        uint   `json:"id"`
	Name      string `json:"name"`
	FavCount  int64  `json:"fav_count"`
	CreatedAt string `json:"created_at"`
}
