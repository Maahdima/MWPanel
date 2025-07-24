package schema

type LoginRequest struct {
	Username string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required"`
}

type LoginResponse struct {
	UserID       uint   `json:"user_id"`
	Username     string `json:"username"`
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"`
}

type UpdateProfileRequest struct {
	OldUsername string  `json:"old_username" validate:"required"`
	OldPassword string  `json:"old_password" validate:"required"`
	NewUsername *string `json:"new_username"`
	NewPassword *string `json:"new_password"`
}
