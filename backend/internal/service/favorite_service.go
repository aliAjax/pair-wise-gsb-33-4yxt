package service

import (
	"errors"
	"fmt"
	"log/slog"

	"github.com/gbplantwiki/gbplantwiki/internal/constants"
	"github.com/gbplantwiki/gbplantwiki/internal/model"
	"github.com/gbplantwiki/gbplantwiki/internal/repository"
	"github.com/gbplantwiki/gbplantwiki/internal/util"
)

// FavoriteService implements favorite/unfavorite/grouping logic.
type FavoriteService struct {
	repo      *repository.FavoriteRepository
	groupRepo *repository.FavoriteGroupRepository
	logger    *slog.Logger
}

// NewFavoriteService creates a FavoriteService.
func NewFavoriteService(repo *repository.FavoriteRepository, groupRepo *repository.FavoriteGroupRepository, logger *slog.Logger) *FavoriteService {
	return &FavoriteService{repo: repo, groupRepo: groupRepo, logger: logger}
}

// Add favorites a target for a user, optionally inside a group.
func (s *FavoriteService) Add(userID uint, targetType string, targetID uint, groupID *uint) (*model.Favorite, error) {
	if !constants.IsValidFavoriteTarget(targetType) {
		return nil, util.NewAppError(422, constants.CodeValidationError,
			fmt.Sprintf("Favorite[target_type=%s] add failed: invalid target type", targetType))
	}
	if groupID != nil && *groupID == 0 {
		groupID = nil
	}
	if groupID != nil {
		if err := s.checkGroupOwner(userID, *groupID); err != nil {
			return nil, err
		}
	}
	f := &model.Favorite{UserID: userID, TargetType: targetType, TargetID: targetID, GroupID: groupID}
	if err := s.repo.Create(f); err != nil {
		if errors.Is(err, repository.ErrDuplicate) {
			return nil, util.NewAppError(409, constants.CodeConflict,
				fmt.Sprintf("Favorite[user_id=%d target=%s:%d] add failed: already favorited", userID, targetType, targetID))
		}
		s.logger.Error("favorite add failed", "error", err)
		return nil, fmt.Errorf("favorite add: %w", err)
	}
	s.logger.Info("favorite added", "user_id", userID, "target_type", targetType, "target_id", targetID, "group_id", groupID)
	return f, nil
}

// Remove deletes a favorite.
func (s *FavoriteService) Remove(userID uint, targetType string, targetID uint) error {
	f, err := s.repo.Find(userID, targetType, targetID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return util.NewAppError(404, constants.CodeNotFound,
				fmt.Sprintf("Favorite[user_id=%d target=%s:%d] not found", userID, targetType, targetID))
		}
		return fmt.Errorf("favorite remove find: %w", err)
	}
	if err := s.repo.Delete(f.ID); err != nil {
		return fmt.Errorf("favorite remove: %w", err)
	}
	s.logger.Info("favorite removed", "user_id", userID, "target_type", targetType, "target_id", targetID)
	return nil
}

// ListFilter selects favorites: groupID nil + uncategorized=true lists "未分类";
// groupID set lists that group; both zero values lists all favorites.
type ListFilter struct {
	GroupID       *uint
	Uncategorized bool
	TargetType    string
}

// ListByUser returns a user's favorites according to the group filter.
func (s *FavoriteService) ListByUser(userID uint, filter ListFilter) ([]model.Favorite, error) {
	if filter.Uncategorized {
		items, err := s.repo.ListUncategorized(userID, filter.TargetType)
		if err != nil {
			return nil, fmt.Errorf("favorite list: %w", err)
		}
		return items, nil
	}
	if filter.GroupID != nil {
		if err := s.checkGroupOwner(userID, *filter.GroupID); err != nil {
			return nil, err
		}
		items, err := s.repo.ListByGroup(userID, *filter.GroupID, filter.TargetType)
		if err != nil {
			return nil, fmt.Errorf("favorite list: %w", err)
		}
		return items, nil
	}
	items, err := s.repo.ListByUser(userID, filter.TargetType)
	if err != nil {
		return nil, fmt.Errorf("favorite list: %w", err)
	}
	return items, nil
}

// Move assigns a favorite to another group; nil/0 moves it back to "未分类".
func (s *FavoriteService) Move(userID uint, targetType string, targetID uint, groupID *uint) (*model.Favorite, error) {
	if groupID != nil && *groupID == 0 {
		groupID = nil
	}
	if groupID != nil {
		if err := s.checkGroupOwner(userID, *groupID); err != nil {
			return nil, err
		}
	}
	f, err := s.repo.Find(userID, targetType, targetID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, util.NewAppError(404, constants.CodeNotFound,
				fmt.Sprintf("Favorite[user_id=%d target=%s:%d] not found", userID, targetType, targetID))
		}
		return nil, fmt.Errorf("favorite move find: %w", err)
	}
	if err := s.repo.SetGroup(f.ID, groupID); err != nil {
		s.logger.Error("favorite move failed", "error", err)
		return nil, fmt.Errorf("favorite move: %w", err)
	}
	f.GroupID = groupID
	s.logger.Info("favorite moved", "user_id", userID, "favorite_id", f.ID, "group_id", groupID)
	return f, nil
}

// checkGroupOwner verifies the target group exists and belongs to the user.
func (s *FavoriteService) checkGroupOwner(userID, groupID uint) error {
	g, err := s.groupRepo.FindByID(groupID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return util.NewAppError(404, constants.CodeNotFound,
				fmt.Sprintf("FavoriteGroup[id=%d] not found", groupID))
		}
		return fmt.Errorf("favorite group find: %w", err)
	}
	if g.UserID != userID {
		return util.NewAppError(403, constants.CodeForbidden,
			fmt.Sprintf("FavoriteGroup[id=%d] access denied: not owner", groupID))
	}
	return nil
}
