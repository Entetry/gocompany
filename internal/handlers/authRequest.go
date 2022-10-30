package handlers

type refreshTokenRequest struct {
	RefreshToken string `json:"refreshToken" validate:"required"`
}

type logoutRequest struct {
	RefreshToken string `json:"refreshToken" validate:"required"`
}

type signInRequest struct {
	Username string `json:"username" validate:"required,gte=3,lte=32"`
	Password string `json:"password" validate:"required,gte=5,lte=60"`
}

type signUpRequest struct {
	Username string `json:"username" validate:"required,gte=3,lte=32"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,gte=5,lte=60"`
}

type tokenResponse struct {
	AccessToken  string
	RefreshToken string
}
