package response

type BasicResponse struct {
	Status  int    `json:"status"`
	Message string `json:"message"`
}
