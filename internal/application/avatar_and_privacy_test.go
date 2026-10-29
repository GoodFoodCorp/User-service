package application

import (
	"bytes"
	"context"
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"goodfood/user-service/internal/domain"
)

type fakeAvatars struct{ objects map[string][]byte }

func newFakeAvatars() *fakeAvatars { return &fakeAvatars{objects: map[string][]byte{}} }

func (f *fakeAvatars) Put(_ context.Context, key, _ string, body io.Reader, _ int64) (string, error) {
	data, _ := io.ReadAll(body)
	url := "https://cdn.test/" + key
	f.objects[url] = data
	return url, nil
}

func (f *fakeAvatars) Delete(_ context.Context, url string) error {
	delete(f.objects, url)
	return nil
}

type fakeEraser struct{ erased []string }

func (f *fakeEraser) EraseUser(_ context.Context, userID string) error {
	f.erased = append(f.erased, userID)
	return nil
}

// A minimal valid PNG header: what http.DetectContentType recognises.
var pngBytes = append([]byte("\x89PNG\r\n\x1a\n"), bytes.Repeat([]byte{0}, 32)...)

func setupPrivacy() (*UseCases, *fakeAvatars, *fakeEraser) {
	avatars, eraser := newFakeAvatars(), &fakeEraser{}
	uc := NewUseCases(Deps{Profiles: newFakeProfiles(), Addresses: newFakeAddresses(), Favorites: newFakeFavorites(),
		Notifications: newFakeNotificationPreferences(), Avatars: avatars, Eraser: eraser})
	return uc, avatars, eraser
}

func TestUploadMyAvatarStoresAndReplaces(t *testing.T) {
	uc, avatars, _ := setupPrivacy()

	first, err := uc.UploadMyAvatar(context.Background(), alice, bytes.NewReader(pngBytes))
	require.NoError(t, err)
	assert.True(t, strings.HasPrefix(first.AvatarURL, "https://cdn.test/avatars/user-1-"))
	assert.True(t, strings.HasSuffix(first.AvatarURL, ".png"))
	firstURL := first.AvatarURL

	second, err := uc.UploadMyAvatar(context.Background(), alice, bytes.NewReader(pngBytes))
	require.NoError(t, err)
	assert.NotEqual(t, firstURL, second.AvatarURL, "a new key per upload, cache-safe")
	assert.NotContains(t, avatars.objects, firstURL, "the previous image is deleted")
	assert.Len(t, avatars.objects, 1)
}

func TestUploadMyAvatarChecksTheBytesNotTheDeclaredType(t *testing.T) {
	uc, _, _ := setupPrivacy()

	_, err := uc.UploadMyAvatar(context.Background(), alice, strings.NewReader("<html><script>alert(1)</script></html>"))
	var derr *domain.Error
	require.ErrorAs(t, err, &derr)
	assert.Equal(t, domain.ErrCodeValidation, derr.Code)

	_, err = uc.UploadMyAvatar(context.Background(), alice, bytes.NewReader(make([]byte, MaxAvatarBytes+1)))
	require.ErrorAs(t, err, &derr)
}

func TestExportMyDataGathersEverything(t *testing.T) {
	uc, _, _ := setupPrivacy()
	_, err := uc.GetMyProfile(context.Background(), alice)
	require.NoError(t, err)

	data, err := uc.ExportMyData(context.Background(), alice)

	require.NoError(t, err)
	assert.Equal(t, alice.UserID, data.Profile.UserID)
	assert.NotNil(t, data.Addresses)
	assert.NotNil(t, data.Favorites)
}

func TestEraseUserDeletesDataAndAvatar(t *testing.T) {
	uc, avatars, eraser := setupPrivacy()
	_, err := uc.UploadMyAvatar(context.Background(), alice, bytes.NewReader(pngBytes))
	require.NoError(t, err)

	require.NoError(t, uc.EraseUser(context.Background(), alice.UserID))

	assert.Empty(t, avatars.objects)
	assert.Equal(t, []string{alice.UserID}, eraser.erased)
	var derr *domain.Error
	require.ErrorAs(t, uc.EraseUser(context.Background(), ""), &derr)
}
