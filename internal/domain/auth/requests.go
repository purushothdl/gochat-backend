package auth

type LoginRequest struct {
    Email    string `json:"email" validate:"required,email"`
    Password string `json:"password" validate:"required"`
}

type RegisterRequest struct {
    Email    string `json:"email" validate:"required,email"`
    Name     string `json:"name" validate:"required,min=2"`
    Password string `json:"password" validate:"required,min=8"`
}

type RefreshTokenRequest struct {
    // Refresh token will come from HTTP-only cookie, not request body
}

type LogoutRequest struct {
    // Can be empty - token comes from cookie
}

type ForgotPasswordRequest struct {
    Email string `json:"email" validate:"required,email"`
}

type ResetPasswordRequest struct {
    Token    string `json:"token" validate:"required"`
    Password string `json:"password" validate:"required,min=8"`
}
