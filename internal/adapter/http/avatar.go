package http

import (
	"net/http"

	"goodfood/user-service/internal/application"
)

// POST /api/users/me/avatar  (multipart/form-data, field "avatar")
func (h *UserHandler) UploadAvatar(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, application.MaxAvatarBytes+64<<10)
	if err := r.ParseMultipartForm(application.MaxAvatarBytes); err != nil {
		writeError(w, r, http.StatusBadRequest, "avatar file is too large (max 5 MB)")
		return
	}
	file, _, err := r.FormFile("avatar")
	if err != nil {
		writeError(w, r, http.StatusBadRequest, "avatar file is required (field 'avatar')")
		return
	}
	defer file.Close()

	profile, err := h.uc.UploadMyAvatar(r.Context(), actorFrom(r), file)
	if err != nil {
		writeDomainError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, toProfileResponse(profile))
}
