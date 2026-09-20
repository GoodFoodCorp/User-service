package http

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"

	"goodfood/user-service/internal/application"
	"goodfood/user-service/internal/domain"
)

// UserHandler exposes profile and address endpoints. No business logic here.
type UserHandler struct {
	uc         *application.UseCases
	uploadsDir string
}

func NewUserHandler(uc *application.UseCases, uploadsDir string) *UserHandler {
	return &UserHandler{uc: uc, uploadsDir: uploadsDir}
}

// POST /internal/profiles  (service-to-service, called by auth-service on register)
func (h *UserHandler) CreateProfileInternal(w http.ResponseWriter, r *http.Request) {
	var body struct {
		UserID string `json:"user_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.UserID == "" {
		writeError(w, r, http.StatusBadRequest, "user_id is required")
		return
	}
	if err := h.uc.EnsureProfile(r.Context(), body.UserID); err != nil {
		writeDomainError(w, r, err)
		return
	}
	writeJSON(w, http.StatusCreated, map[string]string{"status": "created"})
}

// GET /api/users/me
func (h *UserHandler) GetMe(w http.ResponseWriter, r *http.Request) {
	profile, err := h.uc.GetMyProfile(r.Context(), actorFrom(r))
	if err != nil {
		writeDomainError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, toProfileResponse(profile))
}

// PUT /api/users/me
func (h *UserHandler) UpdateMe(w http.ResponseWriter, r *http.Request) {
	var req profileRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, r, http.StatusBadRequest, "invalid JSON body")
		return
	}
	profile, err := h.uc.UpdateMyProfile(r.Context(), actorFrom(r), domain.ProfileInput{
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Phone:     req.Phone,
		Age:       req.Age,
	})
	if err != nil {
		writeDomainError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, toProfileResponse(profile))
}

// GET /api/users/me/addresses
func (h *UserHandler) ListAddresses(w http.ResponseWriter, r *http.Request) {
	list, err := h.uc.ListMyAddresses(r.Context(), actorFrom(r))
	if err != nil {
		writeDomainError(w, r, err)
		return
	}
	out := make([]addressResponse, 0, len(list))
	for i := range list {
		out = append(out, toAddressResponse(&list[i]))
	}
	writeJSON(w, http.StatusOK, out)
}

// POST /api/users/me/addresses
func (h *UserHandler) AddAddress(w http.ResponseWriter, r *http.Request) {
	var req addressRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, r, http.StatusBadRequest, "invalid JSON body")
		return
	}
	address, err := h.uc.AddAddress(r.Context(), actorFrom(r), domain.AddressInput{
		Label:     req.Label,
		Street:    req.Street,
		ZipCode:   req.ZipCode,
		City:      req.City,
		IsDefault: req.IsDefault,
	})
	if err != nil {
		writeDomainError(w, r, err)
		return
	}
	writeJSON(w, http.StatusCreated, toAddressResponse(address))
}

// DELETE /api/users/me/addresses/{id}
func (h *UserHandler) DeleteAddress(w http.ResponseWriter, r *http.Request) {
	if err := h.uc.DeleteAddress(r.Context(), actorFrom(r), chi.URLParam(r, "id")); err != nil {
		writeDomainError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// GET /api/users/me/favorites
func (h *UserHandler) ListFavorites(w http.ResponseWriter, r *http.Request) {
	list, err := h.uc.ListMyFavorites(r.Context(), actorFrom(r))
	if err != nil {
		writeDomainError(w, r, err)
		return
	}
	out := make([]favoriteResponse, 0, len(list))
	for i := range list {
		out = append(out, toFavoriteResponse(&list[i]))
	}
	writeJSON(w, http.StatusOK, out)
}

// POST /api/users/me/favorites
func (h *UserHandler) AddFavorite(w http.ResponseWriter, r *http.Request) {
	var req favoriteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, r, http.StatusBadRequest, "invalid JSON body")
		return
	}
	favorite, err := h.uc.AddFavorite(r.Context(), actorFrom(r), domain.FavoriteKind(req.Kind), req.TargetID)
	if err != nil {
		writeDomainError(w, r, err)
		return
	}
	writeJSON(w, http.StatusCreated, toFavoriteResponse(favorite))
}

// DELETE /api/users/me/favorites/{kind}/{targetId}
func (h *UserHandler) RemoveFavorite(w http.ResponseWriter, r *http.Request) {
	kind := domain.FavoriteKind(chi.URLParam(r, "kind"))
	targetID := chi.URLParam(r, "targetId")
	if err := h.uc.RemoveFavorite(r.Context(), actorFrom(r), kind, targetID); err != nil {
		writeDomainError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// GET /api/users/me/notifications
func (h *UserHandler) GetNotifications(w http.ResponseWriter, r *http.Request) {
	prefs, err := h.uc.GetMyNotificationPreferences(r.Context(), actorFrom(r))
	if err != nil {
		writeDomainError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, toNotificationPreferencesResponse(prefs))
}

// PUT /api/users/me/notifications
func (h *UserHandler) UpdateNotifications(w http.ResponseWriter, r *http.Request) {
	var req notificationPreferencesRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, r, http.StatusBadRequest, "invalid JSON body")
		return
	}
	prefs, err := h.uc.UpdateMyNotificationPreferences(r.Context(), actorFrom(r), domain.NotificationPreferencesInput{
		EmailOrders: req.EmailOrders,
		EmailPromos: req.EmailPromos,
		SmsOrders:   req.SmsOrders,
	})
	if err != nil {
		writeDomainError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, toNotificationPreferencesResponse(prefs))
}
