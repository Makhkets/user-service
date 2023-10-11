package repo

type AboutAccessToken struct {
	Id       int    `json:"id"`
	Username string `json:"username"`
	Status   string `json:"status"`
	IsBanned bool   `json:"isBanned"`
}

type ResponseError struct {
	Error string `json:"error"`
}

type ResponseAccessToken struct {
	Access string `json:"access"`
}

type GenerateTokenForm struct {
	Username string `binding:"required,min=4" json:"username"`
	Password string `binding:"required,min=8" json:"password"`
}

type RefreshTokenForm struct {
	Refresh string `json:"refresh"`
}

type UserDTOForm struct {
	Username string `json:"username" binding:"required,min=4"`
	Password string `json:"password" binding:"required,min=8"`
}

type CreateUserResponseForm struct {
	Access  string `json:"access"`
	Refresh string `json:"refresh"`
}

type MessageResponseForm struct {
	Message string `json:"message"`
}

type GetUserForm struct {
	Id       string `json:"id,omitempty"`
	Username string `json:"username,omitempty"`
	Status   string `json:"status,omitempty"`
	IsBanned *bool  `json:"isBanned,omitempty"`
}

type UserSessionsForm struct {
	Id           string `json:"id,omitempty"`
	Ip           string `json:"ip,omitempty"`
	UserAgent    string `json:"userAgent,omitempty"`
	FingerPrint  string `json:"fingerPrint,omitempty"`
	RefreshToken string `json:"refreshToken,omitempty"`
	CreatedAt    string `json:"createdAt,omitempty"`
	ExpiresIn    string `json:"expiresIn,omitempty"`
}

type UsernameForm struct {
	Username string `json:"username"`
}

type UsernameResponseForm struct {
	NewUsername string `json:"new,omitempty"`
	OldUsername string `json:"old,omitempty"`
}

type PasswordForm struct {
	OldPassword string `json:"old_password" binding:"required,min=8"`
	NewPassword string `json:"new_password" binding:"required,min=8"`
}

type PasswordResponseForm struct {
	NewPassword string `json:"new_password,omitempty"`
	OldPassword string `json:"old_password,omitempty"`
}

type StatusForm struct {
	Status string `json:"status" binding:"required,min=4"`
}

type PermissionForm struct {
	Permission *bool `json:"permission" binding:"required"`
}
