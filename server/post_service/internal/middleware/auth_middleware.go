package middleware

import (
	"context"
	"fmt"
	grpc_client "post_service/internal/clients/grpc"
	"post_service/internal/model/response"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type AuthMiddleware struct {
	GrpcClient *grpc_client.GrpcClient
}

func (a *AuthMiddleware) Run(c *gin.Context) {
	authHeader := c.Request.Header.Get("Authorization")
	if authHeader == "" {
		c.JSON(401, response.BasicResponse{Status: 401, Message: "Unauthorized"})
		c.Abort()
		return
	}
	headerArr := strings.Split(authHeader, " ")
	if len(headerArr) != 2 {
		c.JSON(401, response.BasicResponse{Status: 401, Message: "Unauthorized"})
		c.Abort()
		return
	}
	token := headerArr[1]
	ctx, ctxFunc := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxFunc()
	authResponse, err := a.GrpcClient.Authenticate(ctx, token)
	if err != nil {
		logrus.Errorf(err.Error())
		c.Abort()
		return
	}
	if authResponse.StatusCode != 200 {
		c.JSON(401, response.BasicResponse{Status: 401, Message: "Unauthorized"})
		c.Abort()
		return
	}
	c.AddParam("userId", fmt.Sprintf("%d", authResponse.UserId))
	c.AddParam("username", authResponse.Username)
	c.Next()
}
