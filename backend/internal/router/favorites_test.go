package router

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/gbplantwiki/gbplantwiki/internal/config"
	"github.com/gbplantwiki/gbplantwiki/internal/model"
	"github.com/gbplantwiki/gbplantwiki/internal/util"
)

// TestFavoriteRoutesRequireAuth ensures the favorite grouping endpoints are
// mounted behind the auth middleware.
func TestFavoriteRoutesRequireAuth(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	r := Setup(config.Load(), db, slog.Default())

	paths := [][2]string{
		{http.MethodGet, "/api/v1/favorites"},
		{http.MethodGet, "/api/v1/favorites/groups"},
		{http.MethodPost, "/api/v1/favorites/groups"},
		{http.MethodDelete, "/api/v1/favorites/groups/1"},
		{http.MethodPut, "/api/v1/favorites/plant/1/group"},
		{http.MethodDelete, "/api/v1/favorites/plant/1"},
	}
	for _, p := range paths {
		req := httptest.NewRequest(p[0], p[1], nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code != http.StatusUnauthorized {
			t.Errorf("%s %s: got status %d, want 401", p[0], p[1], w.Code)
		}
	}
}

// envelope mirrors the unified API response.
type envelope struct {
	Code int             `json:"code"`
	Data json.RawMessage `json:"data"`
}

func TestFavoriteGroupHTTPFlow(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	if err := db.AutoMigrate(&model.Favorite{}, &model.FavoriteGroup{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	db.Exec("DELETE FROM favorites")
	db.Exec("DELETE FROM favorite_groups")

	cfg := config.Load()
	cfg.JWTSecret = "router-test-secret"
	r := Setup(cfg, db, slog.Default())

	token, err := util.GenerateToken(7, "tester", "user", cfg.JWTSecret, time.Hour)
	if err != nil {
		t.Fatalf("token: %v", err)
	}
	do := func(method, path, body string) (*httptest.ResponseRecorder, envelope) {
		t.Helper()
		var buf *bytes.Buffer
		if body != "" {
			buf = bytes.NewBufferString(body)
		} else {
			buf = &bytes.Buffer{}
		}
		req := httptest.NewRequest(method, path, buf)
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		var env envelope
		_ = json.Unmarshal(w.Body.Bytes(), &env)
		return w, env
	}

	// create group
	w, env := do(http.MethodPost, "/api/v1/favorites/groups", `{"name":"多肉类"}`)
	if w.Code != http.StatusCreated {
		t.Fatalf("create group: status=%d body=%s", w.Code, w.Body.String())
	}
	var group struct {
		ID uint `json:"id"`
	}
	json.Unmarshal(env.Data, &group)
	if group.ID == 0 {
		t.Fatal("group id missing")
	}

	// duplicate name -> 409
	if w, _ := do(http.MethodPost, "/api/v1/favorites/groups", `{"name":"多肉类"}`); w.Code != http.StatusConflict {
		t.Fatalf("dup group: got %d, want 409", w.Code)
	}

	// add a favorite directly into the group
	body := `{"target_type":"plant","target_id":42,"group_id":` + jsonItoa(group.ID) + `}`
	if w, _ := do(http.MethodPost, "/api/v1/favorites", body); w.Code != http.StatusCreated {
		t.Fatalf("add favorite: got %d, want 201", w.Code)
	}

	// groups list shows count 1 and uncategorized 0
	w, env = do(http.MethodGet, "/api/v1/favorites/groups", "")
	if w.Code != http.StatusOK {
		t.Fatalf("list groups: %d", w.Code)
	}
	var listData struct {
		Groups []struct {
			Name        string `json:"name"`
			FavoriteCnt int64  `json:"favorite_cnt"`
		} `json:"groups"`
		UncategorizedCnt int64 `json:"uncategorized_cnt"`
		TotalFavoriteCnt int64 `json:"total_favorite_cnt"`
	}
	json.Unmarshal(env.Data, &listData)
	if len(listData.Groups) != 1 || listData.Groups[0].FavoriteCnt != 1 ||
		listData.UncategorizedCnt != 0 || listData.TotalFavoriteCnt != 1 {
		t.Fatalf("unexpected group list: %s", env.Data)
	}

	// move favorite back to uncategorized
	w, _ = do(http.MethodPut, "/api/v1/favorites/plant/42/group", `{"group_id":null}`)
	if w.Code != http.StatusOK {
		t.Fatalf("move favorite: %d body=%s", w.Code, w.Body.String())
	}
	w, env = do(http.MethodGet, "/api/v1/favorites/groups", "")
	json.Unmarshal(env.Data, &listData)
	if listData.Groups[0].FavoriteCnt != 0 || listData.UncategorizedCnt != 1 {
		t.Fatalf("counts after move wrong: %s", env.Data)
	}

	// delete group; favorite survives and stays uncategorized
	if w, _ := do(http.MethodDelete, "/api/v1/favorites/groups/"+jsonItoa(group.ID), ""); w.Code != http.StatusOK {
		t.Fatalf("delete group: %d", w.Code)
	}
	w, env = do(http.MethodGet, "/api/v1/favorites?uncategorized=true", "")
	var favs []struct {
		TargetID uint  `json:"target_id"`
		GroupID  *uint `json:"group_id"`
	}
	json.Unmarshal(env.Data, &favs)
	if len(favs) != 1 || favs[0].TargetID != 42 || favs[0].GroupID != nil {
		t.Fatalf("favorite not back in uncategorized: %s", env.Data)
	}
}

func jsonItoa(v uint) string {
	b, _ := json.Marshal(v)
	return string(b)
}
