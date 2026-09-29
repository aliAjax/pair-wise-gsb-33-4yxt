package service

import (
	"errors"
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/gbplantwiki/gbplantwiki/internal/constants"
	"github.com/gbplantwiki/gbplantwiki/internal/model"
	"github.com/gbplantwiki/gbplantwiki/internal/repository"
	"github.com/gbplantwiki/gbplantwiki/internal/util"
)

func newFavoriteTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared&_fk=1"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(&model.Favorite{}, &model.FavoriteGroup{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	t.Cleanup(func() {
		db.Exec("DELETE FROM favorites")
		db.Exec("DELETE FROM favorite_groups")
	})
	return db
}

func setupFavoriteServices(db *gorm.DB) (*FavoriteService, *FavoriteGroupService) {
	log := newTestLogger()
	favRepo := repository.NewFavoriteRepository(db)
	groupRepo := repository.NewFavoriteGroupRepository(db)
	return NewFavoriteService(favRepo, groupRepo, log), NewFavoriteGroupService(groupRepo, log)
}

func appStatus(err error) (int, bool) {
	var ae *util.AppError
	if errors.As(err, &ae) {
		return ae.HTTPStatus, true
	}
	return 0, false
}

func TestFavoriteGroupLifecycle(t *testing.T) {
	db := newFavoriteTestDB(t)
	favSvc, groupSvc := setupFavoriteServices(db)
	const alice, bob uint = 1, 2

	// 1. create groups owned by the current account
	succulents, err := groupSvc.Create(alice, "多肉养护")
	if err != nil {
		t.Fatalf("create group: %v", err)
	}
	articles, err := groupSvc.Create(alice, "文章精选")
	if err != nil {
		t.Fatalf("create second group: %v", err)
	}

	// same account cannot create a duplicate name
	if _, err := groupSvc.Create(alice, " 多肉养护 "); err == nil {
		t.Fatal("duplicate name should be rejected")
	} else if status, ok := appStatus(err); !ok || status != 409 {
		t.Fatalf("expected 409 conflict, got %v", err)
	}

	// the reserved "未分类" name cannot be used by a real group
	if _, err := groupSvc.Create(alice, constants.FavoriteGroupUncategorized); err == nil {
		t.Fatal("reserved name should be rejected")
	}

	// empty/whitespace name rejected
	if _, err := groupSvc.Create(alice, "   "); err == nil {
		t.Fatal("empty name should be rejected")
	}

	// a different account may use the same name
	if _, err := groupSvc.Create(bob, "多肉养护"); err != nil {
		t.Fatalf("other account should reuse name: %v", err)
	}

	// 2. favorites start uncategorized or in a group when added with group_id
	if _, err := favSvc.Add(alice, constants.FavoriteTargetPlant, 10, nil); err != nil {
		t.Fatalf("add uncategorized favorite: %v", err)
	}
	gid := succulents.ID
	if _, err := favSvc.Add(alice, constants.FavoriteTargetPlant, 11, &gid); err != nil {
		t.Fatalf("add grouped favorite: %v", err)
	}
	aid := articles.ID
	if _, err := favSvc.Add(alice, constants.FavoriteTargetArticle, 20, &aid); err != nil {
		t.Fatalf("add article favorite: %v", err)
	}

	// adding with someone else's group is forbidden
	otherGroup := succulents.ID
	if _, err := favSvc.Add(bob, constants.FavoriteTargetPlant, 10, &otherGroup); err == nil {
		t.Fatal("expected forbidden when using other user's group")
	}

	// counts reflect real ownership
	data, err := groupSvc.ListWithCounts(alice)
	if err != nil {
		t.Fatalf("list groups: %v", err)
	}
	if data.UncategorizedCnt != 1 || data.TotalFavoriteCnt != 3 || len(data.Groups) != 2 {
		t.Fatalf("unexpected counts: %+v", data)
	}
	for _, g := range data.Groups {
		if g.FavoriteCnt != 1 {
			t.Errorf("group %s count = %d, want 1", g.Name, g.FavoriteCnt)
		}
	}

	// 3. move a favorite to another group and back to uncategorized
	moved, err := favSvc.Move(alice, constants.FavoriteTargetPlant, 10, &gid)
	if err != nil {
		t.Fatalf("move to group: %v", err)
	}
	if moved.GroupID == nil || *moved.GroupID != gid {
		t.Fatalf("move result group_id = %v, want %d", moved.GroupID, gid)
	}
	if _, err := favSvc.Move(alice, constants.FavoriteTargetPlant, 10, nil); err != nil {
		t.Fatalf("move back to uncategorized: %v", err)
	}

	// 4. list filtering
	all, err := favSvc.ListByUser(alice, ListFilter{})
	if err != nil || len(all) != 3 {
		t.Fatalf("list all favorites: %d %v", len(all), err)
	}
	uncat, err := favSvc.ListByUser(alice, ListFilter{Uncategorized: true})
	if err != nil || len(uncat) != 1 {
		t.Fatalf("list uncategorized: %d %v", len(uncat), err)
	}
	inGroup, err := favSvc.ListByUser(alice, ListFilter{GroupID: &gid})
	if err != nil || len(inGroup) != 1 || inGroup[0].TargetID != 11 {
		t.Fatalf("list by group: %d %v", len(inGroup), err)
	}

	// querying another account's group is forbidden
	if _, err := favSvc.ListByUser(bob, ListFilter{GroupID: &gid}); err == nil {
		t.Fatal("expected forbidden listing other user's group")
	}

	// 5. delete the group: group disappears, favorites return to uncategorized
	if err := groupSvc.Delete(alice, gid); err != nil {
		t.Fatalf("delete group: %v", err)
	}
	var groupCount int64
	db.Model(&model.FavoriteGroup{}).Where("id = ?", gid).Count(&groupCount)
	if groupCount != 0 {
		t.Fatal("group should be gone")
	}
	var stillGrouped int64
	db.Model(&model.Favorite{}).Where("group_id = ?", gid).Count(&stillGrouped)
	if stillGrouped != 0 {
		t.Fatalf("%d favorites still point at deleted group", stillGrouped)
	}
	data, err = groupSvc.ListWithCounts(alice)
	if err != nil {
		t.Fatalf("list groups after delete: %v", err)
	}
	if data.UncategorizedCnt != 2 || data.TotalFavoriteCnt != 3 || len(data.Groups) != 1 {
		t.Fatalf("counts after delete mismatch: %+v", data)
	}
	// the favorite rows survive, so targets still open
	all, _ = favSvc.ListByUser(alice, ListFilter{})
	if len(all) != 3 {
		t.Fatalf("favorites should survive group deletion, got %d", len(all))
	}

	// deleting another user's group is forbidden
	if err := groupSvc.Delete(bob, articles.ID); err == nil {
		t.Fatal("expected forbidden deleting other user's group")
	}
}
