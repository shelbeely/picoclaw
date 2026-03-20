package channels

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/sipeed/picoclaw/pkg/media"
)

type failingMediaStore struct{}

func (failingMediaStore) Store(localPath string, meta media.MediaMeta, scope string) (string, error) {
	return "", os.ErrPermission
}

func (failingMediaStore) Resolve(ref string) (string, error) { return "", os.ErrNotExist }

func (failingMediaStore) ResolveWithMeta(ref string) (string, media.MediaMeta, error) {
	return "", media.MediaMeta{}, os.ErrNotExist
}

func (failingMediaStore) ReleaseAll(scope string) error { return nil }

func TestNewSenderInfo(t *testing.T) {
	sender := NewSenderInfo("discord", "123", "alice", "Alice")
	if sender.Platform != "discord" {
		t.Fatalf("Platform = %q, want %q", sender.Platform, "discord")
	}
	if sender.PlatformID != "123" {
		t.Fatalf("PlatformID = %q, want %q", sender.PlatformID, "123")
	}
	if sender.CanonicalID != "discord:123" {
		t.Fatalf("CanonicalID = %q, want %q", sender.CanonicalID, "discord:123")
	}
	if sender.Username != "alice" {
		t.Fatalf("Username = %q, want %q", sender.Username, "alice")
	}
	if sender.DisplayName != "Alice" {
		t.Fatalf("DisplayName = %q, want %q", sender.DisplayName, "Alice")
	}
}

func TestStoreInboundMediaStoresWithPlatformDefaultSource(t *testing.T) {
	dir := t.TempDir()
	localPath := filepath.Join(dir, "voice.ogg")
	if err := os.WriteFile(localPath, []byte("test"), 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	store := media.NewFileMediaStore()
	ref := StoreInboundMedia(store, "telegram", localPath, media.MediaMeta{
		Filename: "voice.ogg",
	}, "telegram:123:456")

	if ref == localPath {
		t.Fatalf("StoreInboundMedia() returned fallback path, want media ref")
	}

	resolvedPath, resolvedMeta, err := store.ResolveWithMeta(ref)
	if err != nil {
		t.Fatalf("ResolveWithMeta() error = %v", err)
	}
	if resolvedPath != localPath {
		t.Fatalf("Resolved path = %q, want %q", resolvedPath, localPath)
	}
	if resolvedMeta.Source != "telegram" {
		t.Fatalf("Resolved source = %q, want %q", resolvedMeta.Source, "telegram")
	}
}

func TestStoreInboundMediaFallsBackWhenStoreFails(t *testing.T) {
	localPath := filepath.Join(t.TempDir(), "image.png")
	if err := os.WriteFile(localPath, []byte("test"), 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	got := StoreInboundMedia(failingMediaStore{}, "line", localPath, media.MediaMeta{
		Filename: "image.png",
	}, "line:chat:message")

	if got != localPath {
		t.Fatalf("StoreInboundMedia() = %q, want fallback path %q", got, localPath)
	}
}
