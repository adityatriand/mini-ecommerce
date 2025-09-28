package logger

import (
	"fmt"

	"github.com/gin-gonic/gin"
)

const (
	RequestIDKey  = "request_id"
	UserIDKey     = "user_id"
	DefaultValue  = "unknown"
	AnonymousUser = "anonymous"
)

func extractContextValues(c any) (requestID, userID string) {
	requestID = DefaultValue
	userID = AnonymousUser

	ginCtx, ok := c.(*gin.Context)
	if !ok {
		return requestID, userID
	}

	if value, exists := ginCtx.Get(RequestIDKey); exists {
		if strValue, ok := value.(string); ok && strValue != "" {
			requestID = strValue
		}
	}

	if value, exists := ginCtx.Get(UserIDKey); exists {
		switch v := value.(type) {
		case string:
			if v != "" {
				userID = v
			}
		case int:
			userID = fmt.Sprintf("%d", v)
		case uint:
			userID = fmt.Sprintf("%d", v)
		case int64:
			userID = fmt.Sprintf("%d", v)
		case uint64:
			userID = fmt.Sprintf("%d", v)
		default:
			userID = fmt.Sprintf("%v", v)
		}
	}

	return requestID, userID
}
