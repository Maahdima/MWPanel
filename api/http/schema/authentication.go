package schema

type LoginRequest struct {
	Username string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required"`
}

type LoginResponse struct {
	Requires2FA  bool   `json:"requires_2fa"`
	TempToken    string `json:"temp_token,omitempty"`
	UserID       uint   `json:"user_id,omitempty"`
	Username     string `json:"username,omitempty"`
	AccessToken  string `json:"access_token,omitempty"`
	RefreshToken string `json:"refresh_token,omitempty"`
	ExpiresIn    int64  `json:"expires_in,omitempty"`
}

type Verify2FARequest struct {
	TempToken string `json:"temp_token" validate:"required"`
	Code      string `json:"code" validate:"required"`
}

type UpdateProfileRequest struct {
	OldUsername string  `json:"old_username" validate:"required"`
	OldPassword string  `json:"old_password" validate:"required"`
	NewUsername *string `json:"new_username"`
	NewPassword *string `json:"new_password"`
}

type TotpStatusResponse struct {
	Enabled bool `json:"enabled"`
}

type TotpSetupResponse struct {
	Secret        string `json:"secret"`
	OTPAuthURL    string `json:"otpauth_url"`
	QRCodeDataURL string `json:"qr_code_data_url"`
}

type TotpConfirmRequest struct {
	Code string `json:"code" validate:"required"`
}

type TotpConfirmResponse struct {
	RecoveryCodes []string `json:"recovery_codes"`
}

type TotpDisableRequest struct {
	Password string `json:"password" validate:"required"`
	Code     string `json:"code" validate:"required"`
}
