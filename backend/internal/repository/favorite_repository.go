package repository

import (
	"errors"

	"gorm.io/gorm"

	"github.com/gbplantwiki/gbplantwiki/internal/model"
)

// FavoriteRepository handles persistence of favorites.
type FavoriteRepository struct {
	db *gorm.DB
}

// NewFavoriteRepository creates a FavoriteRepository.
func NewFavoriteRepository(db *gorm.DB) *FavoriteRepository {
	return &FavoriteRepository{db: db}
}

// Create inserts a favorite.
func (r *FavoriteRepository) Create(f *model.Favorite) error {
	if err := r.db.Create(f).Error; err != nil {
		if isDuplicate(err) {
			return ErrDuplicate
		}
		return err
	}
	return nil
}

// Find locates a favorite by user and target.
func (r *FavoriteRepository) Find(userID uint, targetType string, targetID uint) (*model.Favorite, error) {
	var f model.Favorite
	if err := r.db.Where("user_id = ? AND target_type = ? AND target_id = ?", userID, targetType, targetID).
		First(&f).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &f, nil
}

// FindByID locates a favorite by its primary key.
func (r *FavoriteRepository) FindByID(id uint) (*model.Favorite, error) {
	var f model.Favorite
	if err := r.db.First(&f, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &f, nil
}

// FindByIDForUser locates a favorite by id and verifies it belongs to the user.
func (r *FavoriteRepository) FindByIDForUser(id, userID uint) (*model.Favorite, error) {
	var f model.Favorite
	if err := r.db.Where("id = ? AND user_id = ?", id, userID).First(&f).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &f, nil
}

// Delete removes a favorite by id.
func (r *FavoriteRepository) Delete(id uint) error {
	return r.db.Delete(&model.Favorite{}, id).Error
}

// ListByUser returns favorites of a user optionally filtered by target type
// and group. A nil groupID applies no group filter; a non-nil groupID filters
// by that group, with 0 meaning uncategorized (group_id IS NULL).
func (r *FavoriteRepository) ListByUser(userID uint, targetType string, groupID *uint) ([]model.Favorite, error) {
	var items []model.Favorite
	q := r.db.Where("user_id = ?", userID)
	if targetType != "" {
		q = q.Where("target_type = ?", targetType)
	}
	if groupID != nil {
		if *groupID == 0 {
			q = q.Where("group_id IS NULL")
		} else {
			q = q.Where("group_id = ?", *groupID)
		}
	}
	if err := q.Order("id DESC").Find(&items).Error; err != nil {
		return nil, err
	}
	return items, nil
}

// UpdateGroup sets the group of a favorite. A nil groupID moves it back to
// uncategorized. The favorite must belong to the acting user.
func (r *FavoriteRepository) UpdateGroup(favoriteID, userID uint, groupID *uint) error {
	q := r.db.Model(&model.Favorite{}).Where("id = ? AND user_id = ?", favoriteID, userID)
	res := q.Update("group_id", groupID)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

// DetachGroupTx clears the group_id of every favorite in a group within a
// transaction, so group deletion moves its favorites back to uncategorized.
func (r *FavoriteRepository) DetachGroupTx(tx *gorm.DB, groupID uint) (int64, error) {
	res := tx.Model(&model.Favorite{}).Where("group_id = ?", groupID).Update("group_id", nil)
	return res.RowsAffected, res.Error
}

// CountByGroups returns favorite counts keyed by group id for a user.
func (r *FavoriteRepository) CountByGroups(userID uint) (map[uint]int64, error) {
	type row struct {
		GroupID *uint
		Count   int64
	}
	var rows []row
	if err := r.db.Model(&model.Favorite{}).
		Select("group_id, COUNT(*) AS count").
		Where("user_id = ?", userID).
		Group("group_id").
		Scan(&rows).Error; err != nil {
		return nil, err
	}
	counts := make(map[uint]int64, len(rows))
	for _, rw := range rows {
		key := uint(0)
		if rw.GroupID != nil {
			key = *rw.GroupID
		}
		counts[key] = rw.Count
	}
	return counts, nil
}
