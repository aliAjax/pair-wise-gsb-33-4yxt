package constants

// FavoriteTargetType enumerates favorite target types.
const (
	FavoriteTargetPlant   = "plant"
	FavoriteTargetArticle = "article"
)

// FavoriteGroup bounds and reserved names.
const (
	FavoriteGroupNameMax = 64
	// FavoriteGroupUncategorized is the virtual bucket for favorites without a group.
	// Real groups may not use this name to avoid ambiguity.
	FavoriteGroupUncategorized = "未分类"
)

// ValidFavoriteTargets returns all accepted target types.
func ValidFavoriteTargets() []string {
	return []string{FavoriteTargetPlant, FavoriteTargetArticle}
}

// IsValidFavoriteTarget reports whether the target type is known.
func IsValidFavoriteTarget(t string) bool {
	for _, v := range ValidFavoriteTargets() {
		if v == t {
			return true
		}
	}
	return false
}
