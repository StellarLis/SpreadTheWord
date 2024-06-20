package request

type UpdatePostRequest struct {
	PostId  int    `json:"post_id"`
	Message string `json:"message"`
}
