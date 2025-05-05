package api

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/starjardin/simplebank/token"
)

const (
	authorizationHeaderKey  = "authorization"
	authorizationTypeBearer = "bearer"
	authorizationPaylaodKey = "authorization_payload"
)

func authMiddleware(tokenMaker token.Maker) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		authorizationHeader := ctx.GetHeader(authorizationHeaderKey)
		if authorizationHeader == "" {
			err := errors.New("authorisation header is not provide")
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, errorResponse(err))
			ctx.Abort()
			return
		}

		fields := strings.Fields(authorizationHeader)

		if len(fields) < 2 {
			err := errors.New("invalid authoraisation header format")
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, errorResponse(err))
			ctx.Abort()
			return
		}

		authorizationHeaderType := strings.ToLower(fields[0])

		if authorizationTypeBearer != authorizationHeaderType {
			err := errors.New("authorization type not supported")
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, errorResponse(err))
			ctx.Abort()
			return
		}

		accessToken := fields[1]

		payload, err := tokenMaker.VerifyToken(accessToken)

		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, errorResponse(err))
			ctx.Abort()
		}

		ctx.Set(authorizationPaylaodKey, payload)
		ctx.Next()
	}
}
