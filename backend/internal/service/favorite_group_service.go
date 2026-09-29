package service

import (
	"errors"
	"fmt"
	"log/slog"
	"strings"

	"github.com/gbplantwiki/gbplantwiki/internal/constants"
	"github.com/gbplantwiki/gbplantwiki/internal/dto"
	"github.com/gbplantwiki/gbplantwiki/internal/model"
	"github.com/gbplantwiki/gbplantwiki/internal/repository"
	"github.com/gbplantwiki/gbplantwiki/internal/util"
	"gorm.io/gorm"
)

// FavoriteGroupService implements personal favorite group management.
// A group belongs to exactly one user; group names are unique per user.
// Deleting a group detaches its favorites back to uncategorized; the
// favorites themselves are never deleted, so their targets keep opening.
type FavoriteGroupService struct {
	db        *gorm.DB
	groupRepo *repository.FavoriteGroupRepository
	favRepo   *repository.FavoriteRepository
	logger    *slog.Logger
}

// NewFavoriteGroupService creates a FavoriteGroupService.
func NewFavoriteGroupService(db *gorm.DB, groupRepo *repository.FavoriteGroupRepository, favRepo *repository.FavoriteRepository, logger *slog.Logger) *FavoriteGroupService {
	return &FavoriteGroupService{db: db, groupRepo: groupRepo, favRepo: favRepo, logger: logger}
}

// Create adds a named group for the current user. Duplicate names are rejected.
func (s *FavoriteGroupService) Create(userID uint, name string) (*model.FavoriteGroup, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, util.NewAppError(422, constants.CodeValidationError,
			"FavoriteGroup create failed: name required")
	}
	if err := s.ensureNameAvailable(userID, name, 0); err != nil {
		return nil, err
	}
	g := &model.FavoriteGroup{UserID: userID, Name: name}
	if err := s.groupRepo.Create(g); err != nil {
		if errors.Is(err, repository.ErrDuplicate) {
			return nil, util.NewAppError(409, constants.CodeConflict,
				fmt.Sprintf("FavoriteGroup[user_id=%d name=%s] create failed: name already exists", userID, name))
		}
		s.logger.Error(fmt.Sprintf(constants.LogFavGroupCreateFailed, name), "error", err)
		return nil, fmt.Errorf("favorite group create: %w", err)
	}
	s.logger.Info(fmt.Sprintf(constants.LogFavGroupCreateSuccess, name), "id", g.ID, "user_id", userID)
	return g, nil
}

// List returns the user's groups with real per-group favorite counts.
func (s *FavoriteGroupService) List(userID uint) ([]dto.FavoriteGroupItem, error) {
	groups, err := s.groupRepo.ListByUser(userID)
	if err != nil {
		return nil, fmt.Errorf("favorite group list: %w", err)
	}
	counts, err := s.favRepo.CountByGroups(userID)
	if err != nil {
		return nil, fmt.Errorf("favorite group count: %w", err)
	}
	items := make([]dto.FavoriteGroupItem, 0, len(groups))
	for _, g := range groups {
		items = append(items, dto.FavoriteGroupItem{
			ID:       g.ID,
			Name:     g.Name,
			FavCount: counts[g.ID],
			CreatedAt: g.CreatedAt.Format("2006-01-02 15:04:05"),
		})
	}
	return items, nil
}

// Rename changes a group's name while keeping the same uniqueness rule.
func (s *FavoriteGroupService) Rename(userID, id uint, name string) (*model.FavoriteGroup, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, util.NewAppError(422, constants.CodeValidationError,
			fmt.Sprintf("FavoriteGroup[id=%d] rename failed: name required", id))
	}
	g, err := s.requireOwnedGroup(userID, id)
	if err != nil {
		return nil, err
	}
	if g.Name == name {
		return g, nil
	}
	if err := s.ensureNameAvailable(userID, name, id); err != nil {
		return nil, err
	}
	if err := s.groupRepo.UpdateName(id, name); err != nil {
		if errors.Is(err, repository.ErrDuplicate) {
			return nil, util.NewAppError(409, constants.CodeConflict,
				fmt.Sprintf("FavoriteGroup[id=%d name=%s] rename failed: name already exists", id, name))
		}
		return nil, fmt.Errorf("favorite group rename: %w", err)
	}
	g.Name = name
	s.logger.Info(fmt.Sprintf(constants.LogFavGroupRenamed, id, name), "user_id", userID)
	return g, nil
}

// Delete removes a group and moves its favorites back to uncategorized in a
// single transaction. Favorited targets are untouched and stay accessible.
func (s *FavoriteGroupService) Delete(userID, id uint) (int64, error) {
	if _, err := s.requireOwnedGroup(userID, id); err != nil {
		return 0, err
	}
	var detached int64
	err := s.db.Transaction(func(tx *gorm.DB) error {
		n, err := s.favRepo.DetachGroupTx(tx, id)
		if err != nil {
			return fmt.Errorf("favorite group delete detach: %w", err)
		}
		detached = n
		if err := s.groupRepo.DeleteTx(tx, id); err != nil {
			return fmt.Errorf("favorite group delete: %w", err)
		}
		return nil
	})
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return 0, util.NewAppError(404, constants.CodeNotFound,
				fmt.Sprintf("FavoriteGroup[id=%d] not found", id))
		}
		return 0, err
	}
	s.logger.Info(fmt.Sprintf(constants.LogFavGroupDeleted, id), "user_id", userID, "detached", detached)
	return detached, nil
}

// requireOwnedGroup loads a group and verifies it belongs to the user.
func (s *FavoriteGroupService) requireOwnedGroup(userID, id uint) (*model.FavoriteGroup, error) {
	g, err := s.groupRepo.FindByID(id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, util.NewAppError(404, constants.CodeNotFound,
				fmt.Sprintf("FavoriteGroup[id=%d] not found", id))
		}
		return nil, fmt.Errorf("favorite group find: %w", err)
	}
	if g.UserID != userID {
		return nil, util.NewAppError(403, constants.CodeForbidden,
			fmt.Sprintf("FavoriteGroup[id=%d] access failed: user_id=%d not owner", id, userID))
	}
	return g, nil
}

// ensureNameAvailable rejects a name already used by another group of the user.
func (s *FavoriteGroupService) ensureNameAvailable(userID uint, name string, excludeID uint) error {
	existing, err := s.groupRepo.FindByName(userID, name)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil
		}
		return fmt.Errorf("favorite group name lookup: %w", err)
	}
	if existing.ID != excludeID {
		return util.NewAppError(409, constants.CodeConflict,
			fmt.Sprintf("FavoriteGroup[user_id=%d name=%s] already exists", userID, name))
	}
	return nil
}
