package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"unipile-connector/internal/domain/errs"
)

// RespondSuccess responds with the appropriate success code and message
func RespondSuccess(c *gin.Context, status int, message string, payload gin.H) {
	body := gin.H{"message": message}
	for k, v := range payload {
		body[k] = v
	}
	c.JSON(status, body)
}

// RespondError responds with the appropriate error code and message
func RespondError(c *gin.Context, err error) {
	if err == nil {
		return
	}

	var codedErr *errs.CodedError
	if errors.As(err, &codedErr) {
		switch codedErr.Kind {
		case errs.ValidationErrorKind:
			c.AbortWithStatusJSON(http.StatusBadRequest, err.(*errs.CodedError))
		case errs.BusinessErrorKind:
			c.AbortWithStatusJSON(http.StatusInternalServerError, err.(*errs.CodedError))
		case errs.SystemErrorKind:
			c.AbortWithStatusJSON(http.StatusInternalServerError, err.(*errs.CodedError))
		}
	} else {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
}

// RespondUnauthorized responds with the appropriate unauthorized code and message
func RespondUnauthorized(c *gin.Context, err error) {
	if err == nil {
		return
	}
	var codedErr *errs.CodedError
	if errors.As(err, &codedErr) {
		c.AbortWithStatusJSON(http.StatusUnauthorized, err.(*errs.CodedError))
	} else {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
	}
}
