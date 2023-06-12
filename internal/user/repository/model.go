package user

import (
	"strings"
	"time"
)

type UserDTO struct {
	Id       string `json:"id"`
	Username string `json:"username" binding:"required,min=4"`
	Password string `json:"password" binding:"required,min=8"`
	IsAdmin  bool   `json:"isAdmin,omitempty"`
	IsBanned bool   `json:"is_banned"`
}

type User struct {
	Id           int       `json:"-"`
	Username     string    `json:"username" binding:"omitempty,min=4"`
	PasswordHash string    `json:"password" binding:"omitempty,min=8"`
	CreatedAt    time.Time `json:"-"`
	UpdatedAt    time.Time `json:"-"`
	IsAdmin      bool      `json:"isAdmin,omitempty"`
	IsBanned     bool      `json:"isBanned,omitempty"`
}

type RefreshSession struct {
	RefreshToken string        `json:"refreshToken"`
	UserId       string        `json:"userId"`
	Ua           string        `json:"ua"`
	Ip           string        `json:"ip"`
	Fingerprint  string        `json:"fingerprint"`
	ExpiresIn    time.Duration `json:"expiresIn"`
	CreatedAt    time.Time     `json:"createdAt"`
}

//id SERIAL PRIMARY KEY,
//username VARCHAR(20) NOT NULL,
//password VARCHAR(50) NOT NULL,
//created_at TIMESTAMP NOT NULL DEFAULT now(),
//updated_at TIMESTAMP NOT NULL DEFAULT now()

var blackListFields []string = []string{
	"Id", "Password", "PasswordHash", "IsAdmin",
	"IsBanned", "CreatedAt", "UpdatedAt",
}

func BlackListCheck(word string) bool {
	for _, field := range blackListFields {
		if strings.ToLower(word) == strings.ToLower(field) {
			return true
		}
	}
	return false
}
