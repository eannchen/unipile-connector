package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"

	"unipile-connector/internal/domain/errs"
)

func TestRespondSuccess(t *testing.T) {
	gin.SetMode(gin.TestMode)
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)

	RespondSuccess(c, http.StatusCreated, "done", gin.H{"data": 42})

	require.Equal(t, http.StatusCreated, rec.Code)
	require.JSONEq(t, `{"message":"done","data":42}`, rec.Body.String())
}

func TestRespondError_WithCodedErrorKinds(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected int
	}{
		{
			name:     "validation",
			err:      errs.WrapValidationError(errors.New("invalid"), "Invalid input"),
			expected: http.StatusBadRequest,
		},
		{
			name:     "business",
			err:      errs.WrapBusinessError(errors.New("oops"), "Failed"),
			expected: http.StatusInternalServerError,
		},
		{
			name:     "system",
			err:      errs.WrapInternalError(errors.New("crash"), "Crashed"),
			expected: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		caseData := tt
		t.Run(caseData.name, func(t *testing.T) {
			gin.SetMode(gin.TestMode)
			rec := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(rec)
			RespondError(c, caseData.err)
			require.Equal(t, caseData.expected, rec.Code)
			var ce errs.CodedError
			require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &ce))
			require.Equal(t, caseData.expected == http.StatusBadRequest, ce.Kind == errs.ValidationErrorKind)
		})
	}
}

func TestRespondError_WithGenericError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)

	RespondError(c, errors.New("boom"))

	require.Equal(t, http.StatusInternalServerError, rec.Code)
	require.JSONEq(t, `{"error":"boom"}`, rec.Body.String())
}

func TestRespondErrorWithNil(t *testing.T) {
	gin.SetMode(gin.TestMode)
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)

	RespondError(c, nil)

	require.Equal(t, http.StatusOK, rec.Code) // no response written
	require.Empty(t, rec.Body.String())
}

func TestRespondUnauthorized(t *testing.T) {
	gin.SetMode(gin.TestMode)
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)

	RespondUnauthorized(c, errs.WrapValidationError(errors.New("bad token"), "Invalid token"))
	require.Equal(t, http.StatusUnauthorized, rec.Code)

	rec = httptest.NewRecorder()
	c, _ = gin.CreateTestContext(rec)
	RespondUnauthorized(c, errors.New("boom"))
	require.Equal(t, http.StatusUnauthorized, rec.Code)
	require.JSONEq(t, `{"error":"boom"}`, rec.Body.String())
}
