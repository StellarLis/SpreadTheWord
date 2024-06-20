package response

import "post_service/internal/model"

type PostResponse struct {
	Status  int          `json:"status"`
	Message string       `json:"message"`
	Post    model.PostDb `json:"post"`
}
