package domain

type UserSettings struct {
	UserID        string `json:"user_id"`
	DarkMode      bool   `json:"dark_mode"`
	Cutlery       bool   `json:"cutlery"`
	OrderTracking bool   `json:"order_tracking"`
	Promos        bool   `json:"promos"`
	Newsletter    bool   `json:"newsletter"`
}

type SettingsInput struct {
	DarkMode      *bool `json:"dark_mode"`
	Cutlery       *bool `json:"cutlery"`
	OrderTracking *bool `json:"order_tracking"`
	Promos        *bool `json:"promos"`
	Newsletter    *bool `json:"newsletter"`
}