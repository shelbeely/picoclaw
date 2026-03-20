package channels

import (
	"github.com/sipeed/picoclaw/pkg/bus"
	"github.com/sipeed/picoclaw/pkg/identity"
	"github.com/sipeed/picoclaw/pkg/media"
)

// PlatformAdapter centralizes the platform-specific normalization that every
// channel performs before handing control back to the shared channel core.
type PlatformAdapter struct {
	platform string
	store    media.MediaStore
}

// NewPlatformAdapter creates a lightweight adapter for a single platform.
func NewPlatformAdapter(platform string, store media.MediaStore) PlatformAdapter {
	return PlatformAdapter{
		platform: platform,
		store:    store,
	}
}

// Sender normalizes platform-specific sender details into the shared SenderInfo shape.
func (a PlatformAdapter) Sender(platformID, username, displayName string) bus.SenderInfo {
	return bus.SenderInfo{
		Platform:    a.platform,
		PlatformID:  platformID,
		CanonicalID: identity.BuildCanonicalID(a.platform, platformID),
		Username:    username,
		DisplayName: displayName,
	}
}

// StoreInboundMedia stores a downloaded inbound asset in the shared media store.
// When no store is configured or storing fails, it falls back to the local path.
func (a PlatformAdapter) StoreInboundMedia(localPath string, meta media.MediaMeta, scope string) string {
	if a.store == nil {
		return localPath
	}
	if meta.Source == "" {
		meta.Source = a.platform
	}

	ref, err := a.store.Store(localPath, meta, scope)
	if err != nil {
		return localPath
	}

	return ref
}

// NewSenderInfo is a convenience wrapper for channels that only need sender normalization.
func NewSenderInfo(platform, platformID, username, displayName string) bus.SenderInfo {
	return NewPlatformAdapter(platform, nil).Sender(platformID, username, displayName)
}

// StoreInboundMedia is a convenience wrapper for channels that only need shared media storage.
func StoreInboundMedia(store media.MediaStore, platform, localPath string, meta media.MediaMeta, scope string) string {
	return NewPlatformAdapter(platform, store).StoreInboundMedia(localPath, meta, scope)
}
