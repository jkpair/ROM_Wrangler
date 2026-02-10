package romdb

import (
	"path/filepath"
	"testing"

	"github.com/kurlmarx/romwrangler/internal/scraper"
	"github.com/kurlmarx/romwrangler/internal/systems"
)

func TestOpenAndClose(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.db")

	db, err := Open(path)
	if err != nil {
		t.Fatalf("Open failed: %v", err)
	}
	defer db.Close()
}

func TestPutAndGet(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.db")

	db, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	info := &scraper.GameInfo{
		Name:      "Super Mario Bros.",
		System:    systems.NintendoNES,
		Region:    "World",
		Publisher: "Nintendo",
		Year:      "1985",
		Source:    "dat",
	}

	sha1 := "facee9c577a5262dbe33b8370e8882c37ea48e2e"

	if err := db.Put(sha1, info); err != nil {
		t.Fatalf("Put failed: %v", err)
	}

	got, ok := db.GetByHash(sha1)
	if !ok {
		t.Fatal("expected to find cached entry")
	}

	if got.Name != info.Name {
		t.Errorf("name = %q, want %q", got.Name, info.Name)
	}
	if got.System != info.System {
		t.Errorf("system = %q, want %q", got.System, info.System)
	}
	if got.Year != info.Year {
		t.Errorf("year = %q, want %q", got.Year, info.Year)
	}
}

func TestGetByHash_NotFound(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.db")

	db, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	_, ok := db.GetByHash("nonexistent")
	if ok {
		t.Error("expected not found")
	}
}
