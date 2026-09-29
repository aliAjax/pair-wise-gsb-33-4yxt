package repository

import (
	"errors"

	"gorm.io/gorm"

	"github.com/gbplantwiki/gbplantwiki/internal/model"
)

// FavoriteGroupRepository handles persistence of favorite groups.
type FavoriteGroupRepository struct {
	db *gorm.DB
}

// NewFavoriteGroupRepository creates a FavoriteGroupRepository.
func NewFavoriteGroupRepository(db *gorm.DB) *FavoriteGroupRepository {
	return &FavoriteGroupRepository{db: db}
}

// Create inserts a favorite group.
func (r *FavoriteGroupRepository) Create(g *model.FavoriteGroup) error {
	if err := r.db.Create(g).Error; err != nil {
		if isDuplicate(err) {
			return ErrDuplicate
		}
		return err
	}
	return nil
}

// FindByID locates a favorite group by primary key.
func (r *FavoriteGroupRepository) FindByID(id uint) (*model.FavoriteGroup, error) {
	var g model.FavoriteGroup
	if err := r.db.First(&g, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &g, nil
}

// FindByName locates a group by owner and name (dup-name check).
func (r *FavoriteGroupRepository) FindByName(userID uint, name string) (*model.FavoriteGroup, error) {
	var g model.FavoriteGroup
	if err := r.db.Where("user_id = ? AND name = ?", userID, name).First(&g).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &g, nil
}

// ListByUser returns all groups owned by a user.
func (r *FavoriteGroupRepository) ListByUser(userID uint) ([]model.FavoriteGroup, error) {
	var items []model.FavoriteGroup
	if err := r.db.Where("user_id = ?", userID).Order("id ASC").Find(&items).Error; err != nil {
		return nil, err
	}
	return items, nil
}

// DeleteAndDetach removes a group and moves its favorites back to "未分类"
// (group_id = NULL) in one transaction. Returns ErrNotFound when the group
// does not exist or is not owned by the user.
func (r *FavoriteGroupRepository) DeleteAndDetach(userID, id uint) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		res := tx.Where("id = ? AND user_id = ?", id, userID).Delete(&model.FavoriteGroup{})
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected == 0 {
			return ErrNotFound
		}
		if err := tx.Model(&model.Favorite{}).
			Where("group_id = ?", id).
			Update("group_id", gorm.Expr("NULL")).Error; err != nil {
			return err
		}
		return nil
	})
}

// FavoriteCountsByGroup returns favorite counts keyed by group id for the
// user's classified favorites.
func (r *FavoriteGroupRepository) FavoriteCountsByGroup(userID uint) (map[uint]int64, error) {
	var rows []struct {
		GroupID uint
		Cnt     int64
	}
	if err := r.db.Model(&model.Favorite{}).
		Select("group_id, COUNT(*) AS cnt").
		Where("user_id = ? AND group_id IS NOT NULL", userID).
		Group("group_id").
		Scan(&rows).Error; err != nil {
		return nil, err
	}
	counts := make(map[uint]int64, len(rows))
	for _, row := range rows {
		counts[row.GroupID] = row.Cnt
	}
	return counts, nil
}

// UncategorizedFavoriteCount counts a user's favorites without a group.
func (r *FavoriteGroupRepository) UncategorizedFavoriteCount(userID uint) (int64, error) {
	var count int64
	if err := r.db.Model(&model.Favorite{}).
		Where("user_id = ? AND group_id IS NULL", userID).
		Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}
