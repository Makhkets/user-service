package repo

type AboutAccessToken struct {
	Id       int    `json:"id"`
	Username string `json:"username"`
	Status   string `json:"status"`
	IsBanned bool   `json:"isBanned"`
}
