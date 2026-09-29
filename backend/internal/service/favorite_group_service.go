package service

import (
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/gbplantwiki/gbplantwiki/internal/constants"
	"github.com/gbplantwiki/gbplantwiki/internal/dto"
	"github.com/gbplantwiki/gbplantwiki/internal/model"
	"github.com/gbplantwiki/gbplantwiki/internal/repository"
	"github.com/gbplantwiki/gbplantwiki/internal/util"
)

// FavoriteGroupService implements personal favorite-group management.
type FavoriteGroupService struct {
	repo   *repository.FavoriteGroupRepository
	logger *slog.Logger
}

// NewFavoriteGroupService creates a FavoriteGroupService.
func NewFavoriteGroupService(repo *repository.FavoriteGroupRepository, logger *slog.Logger) *FavoriteGroupService {
	return &FavoriteGroupService{repo: repo, logger: logger}
}

// Create adds a new group owned by the user; duplicate names are rejected.
func (s *FavoriteGroupService) Create(userID uint, rawName string) (*model.FavoriteGroup, error) {
	name, err := normalizeGroupName(rawName)
	if err != nil {
		return nil, err
	}
	g := &model.FavoriteGroup{UserID: userID, Name: name}
	if err := s.repo.Create(g); err != nil {
		if errors.Is(err, repository.ErrDuplicate) {
			return nil, util.NewAppError(409, constants.CodeConflict,
				fmt.Sprintf("FavoriteGroup[user_id=%d name=%s] create failed: name already exists", userID, name))
		}
		s.logger.Error("favorite group create failed", "error", err)
		return nil, fmt.Errorf("favorite group create: %w", err)
	}
	s.logger.Info("favorite group created", "user_id", userID, "group_id", g.ID, "name", name)
	return g, nil
}

// ListWithCounts returns the user's groups with real favorite counts,
// plus the "未分类" count and the grand total.
func (s *FavoriteGroupService) ListWithCounts(userID uint) (*dto.FavoriteGroupListData, error) {
	groups, err := s.repo.ListByUser(userID)
	if err != nil {
		return nil, fmt.Errorf("favorite group list: %w", err)
	}
	counts, err := s.repo.FavoriteCountsByGroup(userID)
	if err != nil {
		return nil, fmt.Errorf("favorite group counts: %w", err)
	}
	uncategorized, err := s.repo.UncategorizedFavoriteCount(userID)
	if err != nil {
		return nil, fmt.Errorf("favorite group uncategorized count: %w", err)
	}
	items := make([]dto.FavoriteGroupItem, 0, len(groups))
	var total int64 = uncategorized
	for _, g := range groups {
		cnt := counts[g.ID]
		total += cnt
		items = append(items, dto.FavoriteGroupItem{
			ID:          g.ID,
			Name:        g.Name,
			FavoriteCnt: cnt,
			CreatedAt:   g.CreatedAt.Format(time.RFC3339),
		})
	}
	return &dto.FavoriteGroupListData{
		Groups:           items,
		UncategorizedCnt: uncategorized,
		TotalFavoriteCnt: total,
	}, nil
}

// Delete removes a group and sends its favorites back to "未分类".
func (s *FavoriteGroupService) Delete(userID, groupID uint) error {
	g, err := s.repo.FindByID(groupID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return util.NewAppError(404, constants.CodeNotFound,
				fmt.Sprintf("FavoriteGroup[id=%d] not found", groupID))
		}
		return fmt.Errorf("favorite group delete find: %w", err)
	}
	if g.UserID != userID {
		return util.NewAppError(403, constants.CodeForbidden,
			fmt.Sprintf("FavoriteGroup[id=%d] delete denied: not owner", groupID))
	}
	if err := s.repo.DeleteAndDetach(userID, groupID); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return util.NewAppError(404, constants.CodeNotFound,
				fmt.Sprintf("FavoriteGroup[id=%d] not found", groupID))
		}
		s.logger.Error("favorite group delete failed", "error", err)
		return fmt.Errorf("favorite group delete: %w", err)
	}
	s.logger.Info("favorite group deleted; favorites moved to uncategorized", "user_id", userID, "group_id", groupID)
	return nil
}

// normalizeGroupName trims a group name and rejects empty/reserved names.
func normalizeGroupName(name string) (string, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return "", util.NewAppError(422, constants.CodeValidationError, "FavoriteGroup name must not be empty")
	}
	if len([]rune(name)) > constants.FavoriteGroupNameMax {
		return "", util.NewAppError(422, constants.CodeValidationError,
			fmt.Sprintf("FavoriteGroup name exceeds %d characters", constants.FavoriteGroupNameMax))
	}
	if name == constants.FavoriteGroupUncategorized {
		return "", util.NewAppError(422, constants.CodeValidationError,
			fmt.Sprintf("FavoriteGroup name %q is reserved", constants.FavoriteGroupUncategorized))
	}
	return name, nil
}
