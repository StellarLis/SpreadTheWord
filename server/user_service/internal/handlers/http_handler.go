package handlers

import (
	"net/http"
	"strings"
	"time"
	"user_service/internal/metrics"
	"user_service/internal/models"
	"user_service/internal/services"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type HandlerInterface interface {
	New() http.Handler
	Register(c *gin.Context)
	SignIn(c *gin.Context)
	UpdateAvatar(c *gin.Context)
}

type HttpHandler struct {
	UserService services.UserService
}

var _ HandlerInterface = &HttpHandler{}

func (h *HttpHandler) New() http.Handler {
	router := gin.Default()

	userApi := router.Group("/userApi")
	{
		userApi.POST("/register", h.Register)
		userApi.POST("/signin", h.SignIn)
		userApi.PUT("/updateAvatar", h.UpdateAvatar)
	}

	return router.Handler()
}

func (h *HttpHandler) Register(c *gin.Context) {
	startTime := time.Now()
	defer func() {
		metrics.Observe(time.Since(startTime), c.Writer.Status())
	}()

	var user models.UserDto
	err := c.BindJSON(&user)
	if err != nil {
		logrus.Errorln(err)
		c.JSON(http.StatusBadRequest, models.AppError{Message: err.Error()})
		return
	}
	token, err := h.UserService.Register(user.Username, user.Password)
	if err != nil {
		logrus.Errorln(err)
		c.JSON(http.StatusBadRequest, models.AppError{Message: err.Error()})
		return
	}
	c.JSON(http.StatusOK, models.TokenResponse{Token: token})
}

func (h *HttpHandler) SignIn(c *gin.Context) {
	startTime := time.Now()
	defer func() {
		metrics.Observe(time.Since(startTime), c.Writer.Status())
	}()

	var user models.UserDto
	err := c.BindJSON(&user)
	if err != nil {
		logrus.Errorln(err)
		c.Status(http.StatusBadRequest)
		c.JSON(http.StatusBadRequest, models.AppError{Message: err.Error()})
		return
	}
	token, err := h.UserService.SignIn(user.Username, user.Password)
	if err != nil {
		logrus.Errorln(err)
		c.JSON(http.StatusBadRequest, models.AppError{Message: err.Error()})
		return
	}
	c.JSON(http.StatusOK, models.TokenResponse{Token: token})
}

func (h *HttpHandler) UpdateAvatar(c *gin.Context) {
	startTime := time.Now()
	defer func() {
		metrics.Observe(time.Since(startTime), c.Writer.Status())
	}()

	// Getting userId
	authHeader := c.Request.Header.Get("Authorization")
	if authHeader == "" {
		c.JSON(http.StatusUnauthorized, models.AppError{Message: "Unauthorized"})
		return
	}
	headerArr := strings.Split(authHeader, " ")
	if len(headerArr) != 2 {
		c.JSON(http.StatusUnauthorized, models.AppError{Message: "Unauthorized"})
		return
	}
	token := headerArr[1]
	userId, _, err := h.UserService.GetDataFromToken(token)
	if err != nil {
		logrus.Errorln(err)
		c.JSON(http.StatusBadRequest, models.AppError{Message: err.Error()})
		return
	}

	// Getting file
	userFile, _, err := c.Request.FormFile("userFile")
	if err != nil {
		logrus.Errorln(err)
		c.JSON(http.StatusBadRequest, models.AppError{Message: err.Error()})
		return
	}
	defer userFile.Close()
	buff512 := make([]byte, 512)
	if _, err = userFile.Read(buff512); err != nil {
		logrus.Errorln(err)
		c.JSON(http.StatusBadRequest, models.AppError{Message: err.Error()})
		return
	}
	mimeType := http.DetectContentType(buff512)
	if mimeType != "image/png" {
		logrus.Errorln(err)
		c.JSON(http.StatusBadRequest, models.AppError{Message: "Invalid file type"})
		return
	}
	fileBytes := []byte{}
	if _, err = userFile.Read(fileBytes); err != nil {
		logrus.Errorln(err)
		c.JSON(http.StatusBadRequest, models.AppError{Message: "Invalid file type"})
		return
	}

	err = h.UserService.UpdateAvatar(fileBytes, userId)
	if err != nil {
		logrus.Errorln(err)
		c.JSON(http.StatusBadRequest, models.AppError{Message: err.Error()})
		return
	}
	c.JSON(http.StatusOK, models.AppError{Message: "OK"})
}
