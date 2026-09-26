package http

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
)

const maxAvatarBytes = 5 << 20 // 5 MB

var allowedAvatarTypes = map[string]string{
	"image/jpeg": ".jpg",
	"image/png":  ".png",
	"image/webp": ".webp",
}

// POST /api/users/me/avatar  (multipart/form-data, field "avatar")
func (h *UserHandler) UploadAvatar(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, maxAvatarBytes)
	if err := r.ParseMultipartForm(maxAvatarBytes); err != nil {
		writeError(w, r, http.StatusBadRequest, "avatar file is too large (max 5 MB)")
		return
	}
	file, header, err := r.FormFile("avatar")
	if err != nil {
		writeError(w, r, http.StatusBadRequest, "avatar file is required (field 'avatar')")
		return
	}
	defer file.Close()

	ext, ok := allowedAvatarTypes[header.Header.Get("Content-Type")]
	if !ok {
		writeError(w, r, http.StatusBadRequest, "avatar must be a JPEG, PNG or WebP image")
		return
	}

	actor := actorFrom(r)
	if err := os.MkdirAll(h.uploadsDir, 0o755); err != nil {
		writeError(w, r, http.StatusInternalServerError, "could not store avatar")
		return
	}
	filename := actor.UserID + ext
	dst, err := os.Create(filepath.Join(h.uploadsDir, filename))
	if err != nil {
		writeError(w, r, http.StatusInternalServerError, "could not store avatar")
		return
	}
	defer dst.Close()
	if _, err := io.Copy(dst, file); err != nil {
		writeError(w, r, http.StatusInternalServerError, "could not store avatar")
		return
	}

	profile, err := h.uc.SetMyAvatar(r.Context(), actor, fmt.Sprintf("/api/users/avatars/%s", filename))
	if err != nil {
		writeDomainError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, toProfileResponse(profile))
}
