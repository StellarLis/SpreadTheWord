package handler

import (
	"net/http"
	"post_service/internal/metrics"
	"post_service/internal/middleware"
	"post_service/internal/model/request"
	"post_service/internal/model/response"
	"post_service/internal/service"
	"strconv"
	"time"

	grpc_client "post_service/internal/clients/grpc"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type HandlerIn interface {
	New() http.Handler
	GetPost(c *gin.Context)
	NewPost(c *gin.Context)
	UpdatePost(c *gin.Context)
	DeletePost(c *gin.Context)
}

type Handler struct {
	GrpcClient  *grpc_client.GrpcClient
	PostService *service.PostService
}

var _ HandlerIn = &Handler{}

func (h *Handler) New() http.Handler {
	router := gin.Default()
	authMiddleware := middleware.AuthMiddleware{GrpcClient: h.GrpcClient}

	postApi := router.Group("/postApi")
	postApi.Use(authMiddleware.Run)
	{
		postApi.GET("/getPost", h.GetPost)
		postApi.POST("/newPost", h.NewPost)
		postApi.PUT("/updatePost", h.UpdatePost)
		postApi.DELETE("/deletePost", h.DeletePost)
	}

	return router.Handler()
}

func (h *Handler) GetPost(c *gin.Context) {
	const op = "handler.GetPost"

	start := time.Now()
	defer func() {
		metrics.Observe(time.Since(start), c.Writer.Status())
	}()

	postId, err := strconv.ParseInt(c.Query("postId"), 10, 0)
	if err != nil {
		logrus.WithField("op", op).Errorf(err.Error())
		c.JSON(403, response.BasicResponse{Status: 403, Message: "Bad Request"})
		return
	}
	post, err := h.PostService.GetPost(int(postId))
	if err != nil {
		logrus.WithField("op", op).Errorf(err.Error())
		c.JSON(404, response.BasicResponse{Status: 404, Message: "No post with that id was found"})
		return
	}
	c.JSON(200, response.PostResponse{Status: 200, Message: "OK", Post: *post})
}

func (h *Handler) NewPost(c *gin.Context) {
	const op = "handler.NewPost"

	start := time.Now()
	defer func() {
		metrics.Observe(time.Since(start), c.Writer.Status())
	}()

	userId, err := h.getUserId(op, "userId", c)
	if err != nil {
		return
	}
	var request request.NewPostRequest
	err = c.BindJSON(&request)
	if err != nil {
		logrus.WithField("op", op).Errorf(err.Error())
		c.JSON(403, response.BasicResponse{Status: 403, Message: "Bad Request"})
		return
	}

	err = h.PostService.NewPost(request.Message, int(userId))
	if err != nil {
		logrus.Errorf(err.Error())
		c.JSON(403, response.BasicResponse{Status: 403, Message: err.Error()})
		return
	}
	c.JSON(200, response.BasicResponse{Status: 200, Message: "New post created!"})
}

func (h *Handler) UpdatePost(c *gin.Context) {
	const op = "handler.UpdatePost"

	start := time.Now()
	defer func() {
		metrics.Observe(time.Since(start), c.Writer.Status())
	}()

	var request request.UpdatePostRequest
	err := c.BindJSON(&request)
	if err != nil {
		logrus.WithField("op", op).Errorf(err.Error())
		c.JSON(403, response.BasicResponse{Status: 403, Message: "Bad Request"})
		return
	}
	userId, err := h.getUserId(op, "userId", c)
	if err != nil {
		return
	}

	err = h.PostService.UpdatePost(request.PostId, request.Message, int(userId))
	if err != nil {
		logrus.WithField("op", op).Errorf(err.Error())
		c.JSON(403, response.BasicResponse{Status: 403, Message: err.Error()})
		return
	}

	c.JSON(200, response.BasicResponse{Status: 200, Message: "Updated post successfully!"})
}

func (h *Handler) DeletePost(c *gin.Context) {
	const op = "handler.DeletePost"

	start := time.Now()
	defer func() {
		metrics.Observe(time.Since(start), c.Writer.Status())
	}()

	var request request.PostIdRequest
	err := c.BindJSON(&request)
	if err != nil {
		logrus.WithField("op", op).Errorf(err.Error())
		c.JSON(403, response.BasicResponse{Status: 403, Message: "Bad Request"})
		return
	}
	userId, err := h.getUserId(op, "userId", c)
	if err != nil {
		return
	}
	err = h.PostService.DeletePost(request.PostId, userId)
	if err != nil {
		logrus.WithField("op", op).Errorf(err.Error())
		c.JSON(403, response.BasicResponse{Status: 403, Message: err.Error()})
		return
	}
	c.JSON(200, response.BasicResponse{Status: 200, Message: "OK"})
}

func (h *Handler) getUserId(op string, target string, c *gin.Context) (int, error) {
	userId, err := strconv.ParseInt(c.Param(target), 10, 0)
	if err != nil {
		logrus.WithField("op", op).Errorf(err.Error())
		c.JSON(403, response.BasicResponse{Status: 403, Message: "Bad Request"})
		return 0, err
	}

	return int(userId), nil
}
