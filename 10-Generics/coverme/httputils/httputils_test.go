package httputils_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
	"gitlab.com/slon/shad-go/10-Generics/coverme/httputils"
)

func Test_Utils(t *testing.T) {
	w := httptest.NewRecorder()
	require.NoError(t, httputils.RespondJSON(w, 200, ""))

	w2 := httptest.NewRecorder()
	httputils.ServerError(w2)
	require.Equal(t, w2.Code, http.StatusInternalServerError)

	w3 := httptest.NewRecorder()
	httputils.BadRequest(w3, "Bad request")
	require.Equal(t, w3.Code, http.StatusBadRequest)
}
