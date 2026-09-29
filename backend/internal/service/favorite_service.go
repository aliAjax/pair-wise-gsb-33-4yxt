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

// FavoriteService implements favorite/unfavorite logic.
type FavoriteService struct {
	repo      *repository.FavoriteRepository
	groupRepo *repository.FavoriteGroupRepository
	logger    *slog.Logger
}

// NewFavoriteService creates a FavoriteService.
func NewFavoriteService(repo *repository.FavoriteRepository, groupRepo *repository.FavoriteGroupRepository, logger *slog.Logger) *FavoriteService {
	return &FavoriteService{repo: repo, groupRepo: groupRepo, logger: logger}
}

// Add favorites a target for a user, optionally placing it in an owned group.
// A groupID of 0 means uncategorized.
func (s *FavoriteService) Add(userID uint, targetType string, targetID uint, groupID uint) (*model.Favorite, error) {
	if !constants.IsValidFavoriteTarget(targetType) {
		return nil, util.NewAppError(422, constants.CodeValidationError,
			fmt.Sprintf("Favorite[target_type=%s] add failed: invalid target type", targetType))
	}
	var gid *uint
	if groupID != 0 {
		g, err := s.groupRepo.FindByID(groupID)
		if err != nil {
			if errors.Is(err, repository.ErrNotFound) {
				return nil, util.NewAppError(404, constants.CodeNotFound,
					fmt.Sprintf("FavoriteGroup[id=%d] not found", groupID))
			}
			return nil, fmt.Errorf("favorite group find: %w", err)
		}
		if g.UserID != userID {
			return nil, util.NewAppError(403, constants.CodeForbidden,
				fmt.Sprintf("FavoriteGroup[id=%d] add failed: user_id=%d not owner", groupID, userID))
		}
		gid = &groupID
	}
	f := &model.Favorite{UserID: userID, TargetType: targetType, TargetID: targetID, GroupID: gid}
	if err := s.repo.Create(f); err != nil {
		if errors.Is(err, repository.ErrDuplicate) {
			return nil, util.NewAppError(409, constants.CodeConflict,
				fmt.Sprintf("Favorite[user_id=%d target=%s:%d] add failed: already favorited", userID, targetType, targetID))
		}
		s.logger.Error("favorite add failed", "error", err)
		return nil, fmt.Errorf("favorite add: %w", err)
	}
	s.logger.Info("favorite added", "user_id", userID, "target_type", targetType, "target_id", targetID, "group_id", gid)
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

// ListByUser returns a user's favorites, optionally filtered by target type
// and group. groupID nil = all groups; 0 = uncategorized; >0 = that group.
func (s *FavoriteService) ListByUser(userID uint, targetType string, groupID *uint) ([]model.Favorite, error) {
	items, err := s.repo.ListByUser(userID, targetType, groupID)
	if err != nil {
		return nil, fmt.Errorf("favorite list: %w", err)
	}
	return items, nil
}

// Move assigns an existing favorite to another group owned by the user.
// groupID 0 moves it back to uncategorized.
func (s *FavoriteService) Move(userID, favoriteID, groupID uint) (*model.Favorite, error) {
	if groupID != 0 {
		g, err := s.groupRepo.FindByID(groupID)
		if err != nil {
			if errors.Is(err, repository.ErrNotFound) {
				return nil, util.NewAppError(404, constants.CodeNotFound,
					fmt.Sprintf("FavoriteGroup[id=%d] not found", groupID))
			}
			return nil, fmt.Errorf("favorite move group find: %w", err)
		}
		if g.UserID != userID {
			return nil, util.NewAppError(403, constants.CodeForbidden,
				fmt.Sprintf("FavoriteGroup[id=%d] move failed: user_id=%d not owner", groupID, userID))
		}
	}
	var gid *uint
	if groupID != 0 {
		gid = &groupID
	}
	if err := s.repo.UpdateGroup(favoriteID, userID, gid); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, util.NewAppError(404, constants.CodeNotFound,
				fmt.Sprintf("Favorite[id=%d] not found", favoriteID))
		}
		return nil, fmt.Errorf("favorite move: %w", err)
	}
	f, err := s.repo.FindByIDForUser(favoriteID, userID)
	if err != nil {
		return nil, fmt.Errorf("favorite move reload: %w", err)
	}
	s.logger.Info("favorite moved", "id", favoriteID, "user_id", userID, "group_id", gid)
	return f, nil
}
