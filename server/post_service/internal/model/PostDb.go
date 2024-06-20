package model

type PostDb struct {
	PostId   int    `json:"post_id"`
	Message  string `json:"message"`
	UserId   int    `json:"user_id"`
	Username string `json:"username"`
	Avatar   []byte `json:"avatar"`
}
