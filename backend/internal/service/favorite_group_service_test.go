package service

import (
	"errors"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"gorm.io/gorm"

	"github.com/gbplantwiki/gbplantwiki/internal/repository"
)

func newGroupServiceDB(t *testing.T) (*gorm.DB, sqlmock.Sqlmock, *FavoriteGroupService) {
	t.Helper()
	db, mock := newServiceDB(t)
	favRepo := repository.NewFavoriteRepository(db)
	groupRepo := repository.NewFavoriteGroupRepository(db)
	svc := NewFavoriteGroupService(db, groupRepo, favRepo, newTestLogger())
	return db, mock, svc
}

func TestFavoriteGroupCreateRejectsEmptyName(t *testing.T) {
	_, _, svc := newGroupServiceDB(t)
	if _, err := svc.Create(1, "   "); err == nil {
		t.Error("expected validation error for empty group name")
	}
}

func TestFavoriteGroupCreateRejectsDuplicateName(t *testing.T) {
	_, mock, svc := newGroupServiceDB(t)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `favorite_groups` WHERE user_id = ? AND name = ? ORDER BY `favorite_groups`.`id` LIMIT ?")).
		WithArgs(uint(1), "多肉", 1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "name"}).
			AddRow(7, 1, "多肉"))
	_, err := svc.Create(1, "多肉")
	if err == nil {
		t.Fatal("expected conflict error for duplicate name")
	}
}

func TestFavoriteGroupCreateForbidsForeignNameLookupError(t *testing.T) {
	_, mock, svc := newGroupServiceDB(t)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `favorite_groups`")).
		WillReturnError(errors.New("db down"))
	if _, err := svc.Create(1, "月季"); err == nil {
		t.Error("expected error when name lookup fails")
	}
}

func TestFavoriteGroupCreateSuccess(t *testing.T) {
	_, mock, svc := newGroupServiceDB(t)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `favorite_groups` WHERE user_id = ? AND name = ? ORDER BY `favorite_groups`.`id` LIMIT ?")).
		WithArgs(uint(1), "多肉", 1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "name"}))
	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO `favorite_groups`")).
		WillReturnResult(sqlmock.NewResult(11, 1))
	mock.ExpectCommit()
	g, err := svc.Create(1, "多肉")
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if g.ID != 11 || g.UserID != 1 || g.Name != "多肉" {
		t.Errorf("unexpected group: %+v", g)
	}
}

func TestFavoriteGroupDeleteForbiddenForNonOwner(t *testing.T) {
	_, mock, svc := newGroupServiceDB(t)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `favorite_groups` WHERE `favorite_groups`.`id` = ? ORDER BY `favorite_groups`.`id` LIMIT ?")).
		WithArgs(uint(7), 1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "name"}).
			AddRow(7, 2, "别人的分组"))
	if _, err := svc.Delete(1, 7); err == nil {
		t.Error("expected forbidden error when deleting another user's group")
	}
}

func TestFavoriteGroupDeleteDetachesThenRemoves(t *testing.T) {
	_, mock, svc := newGroupServiceDB(t)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `favorite_groups` WHERE `favorite_groups`.`id` = ? ORDER BY `favorite_groups`.`id` LIMIT ?")).
		WithArgs(uint(7), 1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "name"}).
			AddRow(7, 1, "多肉"))
	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta("UPDATE `favorites` SET `group_id`=NULL WHERE group_id = ?")).
		WithArgs(uint(7)).
		WillReturnResult(sqlmock.NewResult(0, 3))
	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM `favorite_groups` WHERE `favorite_groups`.`id` = ?")).
		WithArgs(uint(7)).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()
	detached, err := svc.Delete(1, 7)
	if err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if detached != 3 {
		t.Errorf("expected 3 detached favorites, got %d", detached)
	}
}
