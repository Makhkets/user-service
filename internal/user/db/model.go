package user

type UserDTO struct {
	Id       string `json:"id"`
	Username string `json:"username" binding:"required,min=4"`
	Password string `json:"password" binding:"required,min=8"`
	IsAdmin  bool   `json:"isAdmin,omitempty"`
}

type User struct {
	Id       int
	Username string
	Password string
	IsAdmin  bool
}

//id SERIAL PRIMARY KEY,
//username VARCHAR(20) NOT NULL,
//password VARCHAR(50) NOT NULL,
//created_at TIMESTAMP NOT NULL DEFAULT now(),
//updated_at TIMESTAMP NOT NULL DEFAULT now()
