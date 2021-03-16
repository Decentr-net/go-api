package api

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/Decentr-net/cerberus/pkg/api"

	"github.com/Decentr-net/go-api/test"
)

func Test_WriteOK(t *testing.T) {
	w := httptest.NewRecorder()
	WriteOK(w, http.StatusCreated, struct {
		M int
		N string
	}{
		M: 5,
		N: "str",
	})

	assert.Equal(t, http.StatusCreated, w.Code)
	assert.JSONEq(t, `{"M":5,"N":"str"}`, w.Body.String())
}

func Test_WriteError(t *testing.T) {
	w := httptest.NewRecorder()
	WriteError(w, http.StatusNotFound, "some error")

	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.JSONEq(t, `{"error":"some error"}`, w.Body.String())
}

func Test_WriteErrorf(t *testing.T) {
	w := httptest.NewRecorder()
	WriteErrorf(w, http.StatusForbidden, "some error %d", 1)

	assert.Equal(t, http.StatusForbidden, w.Code)
	assert.JSONEq(t, `{"error":"some error 1"}`, w.Body.String())
}

func Test_WriteInternalError(t *testing.T) {
	b, w, r := test.NewAPITestParameters(http.MethodGet, "", nil)

	WriteInternalError(r.Context(), w, "some error")

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Greater(t, len(b.String()), 20) // stacktrace
	assert.True(t, strings.Contains(b.String(), "some error"))
	assert.JSONEq(t, `{"error":"internal error"}`, w.Body.String())
}

func Test_WriteVerifyError(t *testing.T) {
	t.Run("bad request", func(t *testing.T) {
		b, w, r := test.NewAPITestParameters(http.MethodGet, "", nil)

		WriteVerifyError(r.Context(), w, api.ErrInvalidPublicKey)

		assert.Empty(t, b.String())
		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.JSONEq(t, `{"error":"invalid request: public key is invalid"}`, w.Body.String())
	})

	t.Run("not verified", func(t *testing.T) {
		b, w, r := test.NewAPITestParameters(http.MethodGet, "", nil)
		WriteVerifyError(r.Context(), w, api.ErrNotVerified)

		assert.Empty(t, b.String())
		assert.Equal(t, http.StatusUnauthorized, w.Code)
		assert.JSONEq(t, `{"error":"failed to verify message"}`, w.Body.String())
	})

	t.Run("internal error", func(t *testing.T) {
		b, w, r := test.NewAPITestParameters(http.MethodGet, "", nil)

		WriteVerifyError(r.Context(), w, errors.New("some error"))

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Greater(t, len(b.String()), 20) // stacktrace
		assert.True(t, strings.Contains(b.String(), "some error"))
		assert.JSONEq(t, `{"error":"internal error"}`, w.Body.String())
	})
}
