package httpx

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
)

type Error struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Status  int    `json:"-"`
}

func (e *Error) Error() string { return e.Message }

func NewError(status int, code, message string) *Error {
	return &Error{Status: status, Code: code, Message: message}
}

func OK(c *gin.Context, status int, data any) {
	c.JSON(status, gin.H{"data": data})
}

func Fail(c *gin.Context, err error) {
	var apiErr *Error
	if errors.As(err, &apiErr) {
		c.JSON(apiErr.Status, gin.H{"error": apiErr})
		return
	}

	c.JSON(http.StatusInternalServerError, gin.H{
		"error": Error{
			Code:    "internal_error",
			Message: "internal server error",
		},
	})
}

