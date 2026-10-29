package application

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"io"
	"net/http"

	zlog "github.com/rs/zerolog/log"

	"goodfood/user-service/internal/domain"
)

// MaxAvatarBytes bounds an avatar upload.
const MaxAvatarBytes = 5 << 20

// avatarTypes maps the image types accepted, detected from the bytes
// themselves (the declared content type is not trusted), to an extension.
var avatarTypes = map[string]string{
	"image/jpeg": ".jpg",
	"image/png":  ".png",
	"image/webp": ".webp",
}

// UploadMyAvatar stores a new avatar and replaces the previous one.
func (uc *UseCases) UploadMyAvatar(ctx context.Context, actor Actor, body io.Reader) (*domain.Profile, error) {
	data, err := io.ReadAll(io.LimitReader(body, MaxAvatarBytes+1))
	if err != nil {
		return nil, err
	}
	if len(data) == 0 || len(data) > MaxAvatarBytes {
		return nil, domain.NewValidationError("avatar must be a non-empty image of at most 5 MB")
	}
	contentType := http.DetectContentType(data)
	ext, ok := avatarTypes[contentType]
	if !ok {
		return nil, domain.NewValidationError("avatar must be a JPEG, PNG or WebP image")
	}
	profile, err := uc.GetMyProfile(ctx, actor)
	if err != nil {
		return nil, err
	}

	// A fresh key per upload: CDN and browser caches never serve a stale image.
	key := "avatars/" + actor.UserID + "-" + randomSuffix() + ext
	url, err := uc.avatars.Put(ctx, key, contentType, bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return nil, err
	}
	previous := profile.AvatarURL
	profile.SetAvatar(url)
	if err := uc.profiles.Update(ctx, profile); err != nil {
		_ = uc.avatars.Delete(ctx, url)
		return nil, err
	}
	if previous != "" {
		if err := uc.avatars.Delete(ctx, previous); err != nil {
			zlog.Warn().Err(err).Str("user_id", actor.UserID).Msg("previous avatar not deleted")
		}
	}
	return profile, nil
}

// PersonalData is everything this service stores about a user (right of
// access / portability).
type PersonalData struct {
	Profile       *domain.Profile
	Addresses     []domain.Address
	Favorites     []domain.Favorite
	Settings      *domain.UserSettings
	Notifications *domain.NotificationPreferences
}

// ExportMyData gathers the caller's data.
func (uc *UseCases) ExportMyData(ctx context.Context, actor Actor) (*PersonalData, error) {
	profile, err := uc.GetMyProfile(ctx, actor)
	if err != nil {
		return nil, err
	}
	out := &PersonalData{Profile: profile}
	if out.Addresses, err = uc.addresses.ListByUser(ctx, actor.UserID); err != nil {
		return nil, err
	}
	if out.Favorites, err = uc.favorites.ListByUser(ctx, actor.UserID); err != nil {
		return nil, err
	}
	if uc.settings != nil {
		if out.Settings, err = uc.settings.GetSettings(ctx, actor.UserID); err != nil {
			return nil, err
		}
	}
	if out.Notifications, err = uc.notifications.GetByUserID(ctx, actor.UserID); err != nil && !isNotFound(err) {
		return nil, err
	}
	return out, nil
}

// EraseUser deletes every piece of data about a user, avatar included
// (right to erasure). Idempotent. Called by the user, or by auth-service when
// the account is deleted.
func (uc *UseCases) EraseUser(ctx context.Context, userID string) error {
	if userID == "" {
		return domain.NewValidationError("user id is required")
	}
	if profile, err := uc.profiles.GetByUserID(ctx, userID); err == nil && profile.AvatarURL != "" {
		if err := uc.avatars.Delete(ctx, profile.AvatarURL); err != nil {
			return err
		}
	}
	return uc.eraser.EraseUser(ctx, userID)
}

func isNotFound(err error) bool {
	var derr *domain.Error
	return asDomainError(err, &derr) && derr.Code == domain.ErrCodeNotFound
}

func randomSuffix() string {
	b := make([]byte, 6)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}
