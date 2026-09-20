package http

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"goodfood/user-service/internal/domain"
)

type profileRequest struct {
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Phone     string `json:"phone"`
	Age       *int   `json:"age"`
}

type profileResponse struct {
	UserID    string    `json:"user_id"`
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
	Phone     string    `json:"phone"`
	AvatarURL string    `json:"avatar_url"`
	Age       *int      `json:"age"`
	UpdatedAt time.Time `json:"updated_at"`
}

type favoriteRequest struct {
	Kind     string `json:"kind"`
	TargetID string `json:"target_id"`
}

type favoriteResponse struct {
	ID        string    `json:"id"`
	Kind      string    `json:"kind"`
	TargetID  string    `json:"target_id"`
	CreatedAt time.Time `json:"created_at"`
}

type notificationPreferencesRequest struct {
	EmailOrders bool `json:"email_orders"`
	EmailPromos bool `json:"email_promos"`
	SmsOrders   bool `json:"sms_orders"`
}

type notificationPreferencesResponse struct {
	EmailOrders bool      `json:"email_orders"`
	EmailPromos bool      `json:"email_promos"`
	SmsOrders   bool      `json:"sms_orders"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type addressRequest struct {
	Label     string `json:"label"`
	Street    string `json:"street"`
	ZipCode   string `json:"zip_code"`
	City      string `json:"city"`
	IsDefault bool   `json:"is_default"`
}

type addressResponse struct {
	ID        string `json:"id"`
	Label     string `json:"label"`
	Street    string `json:"street"`
	ZipCode   string `json:"zip_code"`
	City      string `json:"city"`
	IsDefault bool   `json:"is_default"`
	Full      string `json:"full_address"`
}

type errorResponse struct {
	Error     string `json:"error"`
	RequestID string `json:"request_id,omitempty"`
}

func toProfileResponse(p *domain.Profile) profileResponse {
	return profileResponse{
		UserID:    p.UserID,
		FirstName: p.FirstName,
		LastName:  p.LastName,
		Phone:     p.Phone,
		AvatarURL: p.AvatarURL,
		Age:       p.Age,
		UpdatedAt: p.UpdatedAt,
	}
}

func toFavoriteResponse(f *domain.Favorite) favoriteResponse {
	return favoriteResponse{ID: f.ID, Kind: string(f.Kind), TargetID: f.TargetID, CreatedAt: f.CreatedAt}
}

func toNotificationPreferencesResponse(p *domain.NotificationPreferences) notificationPreferencesResponse {
	return notificationPreferencesResponse{
		EmailOrders: p.EmailOrders,
		EmailPromos: p.EmailPromos,
		SmsOrders:   p.SmsOrders,
		UpdatedAt:   p.UpdatedAt,
	}
}

func toAddressResponse(a *domain.Address) addressResponse {
	return addressResponse{
		ID:        a.ID,
		Label:     a.Label,
		Street:    a.Street,
		ZipCode:   a.ZipCode,
		City:      a.City,
		IsDefault: a.IsDefault,
		Full:      a.FullAddress(),
	}
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func writeError(w http.ResponseWriter, r *http.Request, status int, msg string) {
	reqID, _ := r.Context().Value(ctxKeyRequestID).(string)
	writeJSON(w, status, errorResponse{Error: msg, RequestID: reqID})
}

func writeDomainError(w http.ResponseWriter, r *http.Request, err error) {
	var derr *domain.Error
	if errors.As(err, &derr) {
		status := map[domain.ErrorCode]int{
			domain.ErrCodeValidation: http.StatusBadRequest,
			domain.ErrCodeNotFound:   http.StatusNotFound,
			domain.ErrCodeForbidden:  http.StatusForbidden,
		}[derr.Code]
		if status == 0 {
			status = http.StatusInternalServerError
		}
		writeError(w, r, status, derr.Message)
		return
	}
	writeError(w, r, http.StatusInternalServerError, "internal server error")
}
