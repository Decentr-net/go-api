package api

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	logging "github.com/Decentr-net/logrus/context"

	"github.com/Decentr-net/go-api/test"
)

func Test_RecovererMiddleware(t *testing.T) {
	b, w, r := test.NewAPITestParameters(http.MethodGet, "", nil)

	require.NotPanics(t, func() {
		RecovererMiddleware(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
			panic("some panic")
		})).ServeHTTP(w, r)
	})

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Equal(t, `{"error":"internal error"}`, w.Body.String())
	assert.Contains(t, b.String(), "some panic")
}

func Test_LoggerMiddleware(t *testing.T) {
	w := httptest.NewRecorder()
	r, err := http.NewRequest(http.MethodPost, "", nil)
	require.NoError(t, err)

	LoggerMiddleware(http.HandlerFunc(func(_ http.ResponseWriter, ir *http.Request) {
		l := logging.GetLogger(ir.Context())
		assert.NotNil(t, l)
	})).ServeHTTP(w, r)
}

func Test_RequestIDMiddleware(t *testing.T) {
	l, w, r := test.NewAPITestParameters(http.MethodGet, "/", nil)

	var id string

	RequestIDMiddleware(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		id = w.Header().Get("X-Request-ID")

		logging.GetLogger(r.Context()).Info("hi")
	})).ServeHTTP(w, r)

	assert.NotEmpty(t, id)
	assert.Contains(t, l.String(), fmt.Sprintf("request_id=%s", id))
}

func Test_TimeoutMiddleware(t *testing.T) {
	l, w, r := test.NewAPITestParameters(http.MethodGet, "/", nil)
	TimeoutMiddleware(time.Millisecond*5)(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		ctx, _ := context.WithTimeout(context.Background(), time.Second)
		select {
		case <-r.Context().Done():
			require.True(t, errors.Is(r.Context().Err(), context.DeadlineExceeded))
		case <-ctx.Done():
			assert.Fail(t, "should be timed out")
		}
	})).ServeHTTP(w, r)

	s := regexp.MustCompile(`elapsed_time="?(.+)"?`).FindStringSubmatch(l.String())
	require.Len(t, s, 2)

	tt, err := time.ParseDuration(s[1])
	require.NoError(t, err)
	require.NotZero(t, tt.Milliseconds())
}

func Test_BodyLimiterMiddleware(t *testing.T) {
	w := httptest.NewRecorder()
	r, err := http.NewRequest(http.MethodPost, "", bytes.NewReader(make([]byte, 10000)))
	require.NoError(t, err)

	BodyLimiterMiddleware(1000)(http.HandlerFunc(func(_ http.ResponseWriter, ir *http.Request) {
		_, err := ioutil.ReadAll(ir.Body)
		assert.Error(t, err)
		assert.Equal(t, "http: request body too large", err.Error())
	})).ServeHTTP(w, r)
}
