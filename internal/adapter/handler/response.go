package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"unipile-connector/internal/domain/errs"
)

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
