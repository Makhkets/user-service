package storage

type UserDTO struct {
	Username string `json:"username" binding:"required,min=4"`
	Password string `json:"password" binding:"required,min=8"`
	IsAdmin  bool   `json:"isAdmin,omitempty"`
}

//id SERIAL PRIMARY KEY,
//username VARCHAR(20) NOT NULL,
//password VARCHAR(50) NOT NULL,
//created_at TIMESTAMP NOT NULL DEFAULT now(),
//updated_at TIMESTAMP NOT NULL DEFAULT now()
