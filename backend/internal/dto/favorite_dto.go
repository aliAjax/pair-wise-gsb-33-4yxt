package dto

// FavoriteRequest is the payload for adding a favorite.
type FavoriteRequest struct {
	TargetType string `json:"target_type" binding:"required"`
	TargetID   uint   `json:"target_id" binding:"required"`
	// GroupID is optional; omitted/0 puts the favorite into "未分类".
	GroupID *uint `json:"group_id"`
}

// FavoriteMoveRequest reassigns a favorite to another group.
// A missing group_id or 0 moves the favorite back to "未分类".
type FavoriteMoveRequest struct {
	GroupID *uint `json:"group_id"`
}

// FavoriteGroupCreateRequest creates a personal favorite group.
type FavoriteGroupCreateRequest struct {
	Name string `json:"name" binding:"required,max=64"`
}

// FavoriteGroupItem is a group with the real number of favorites it contains.
type FavoriteGroupItem struct {
	ID          uint   `json:"id"`
	Name        string `json:"name"`
	FavoriteCnt int64  `json:"favorite_cnt"`
	CreatedAt   string `json:"created_at"`
}

// FavoriteGroupListData is the response of GET /favorites/groups.
type FavoriteGroupListData struct {
	Groups           []FavoriteGroupItem `json:"groups"`
	UncategorizedCnt int64               `json:"uncategorized_cnt"`
	TotalFavoriteCnt int64               `json:"total_favorite_cnt"`
}
