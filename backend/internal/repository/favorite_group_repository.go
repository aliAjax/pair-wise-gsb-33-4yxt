package repository

import (
	"errors"

	"gorm.io/gorm"

	"github.com/gbplantwiki/gbplantwiki/internal/model"
)

// FavoriteGroupRepository handles persistence of personal favorite groups.
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

// FindByID locates a group by id.
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

// FindByName locates a group by user id and name.
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

// UpdateName renames a group.
func (r *FavoriteGroupRepository) UpdateName(id uint, name string) error {
	res := r.db.Model(&model.FavoriteGroup{}).Where("id = ?", id).Update("name", name)
	if res.Error != nil {
		if isDuplicate(res.Error) {
			return ErrDuplicate
		}
		return res.Error
	}
	if res.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

// DeleteTx removes a group within a transaction.
func (r *FavoriteGroupRepository) DeleteTx(tx *gorm.DB, id uint) error {
	res := tx.Delete(&model.FavoriteGroup{}, id)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

// ListByUser returns all groups of a user ordered by creation time.
func (r *FavoriteGroupRepository) ListByUser(userID uint) ([]model.FavoriteGroup, error) {
	var groups []model.FavoriteGroup
	if err := r.db.Where("user_id = ?", userID).Order("id ASC").Find(&groups).Error; err != nil {
		return nil, err
	}
	return groups, nil
}
