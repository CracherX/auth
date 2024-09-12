package dto

// AccessRequest DTO структура запроса Access операции.
type AccessRequest struct {
	GUID string `validate:"required,uuid"`
	IP   string `validate:"required,ip"`
}

// RefreshRequest DTO структура запроса Refresh операции.
type RefreshRequest struct {
	RefreshToken string `json:"refreshToken" validate:"required"`
	AccessToken  string `json:"accessToken" validate:"required"`
	IP           string `json:"ip" validate:"required,ip"`
}
