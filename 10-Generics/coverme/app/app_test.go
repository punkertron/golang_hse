package app

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"gitlab.com/slon/shad-go/10-Generics/coverme/models"
)

type testStorage struct{}

func (_ *testStorage) AddTodo(string, string) (*models.Todo, error) {
	return nil, errors.New("Add")
}

func (_ *testStorage) GetTodo(models.ID) (*models.Todo, error) {
	return nil, errors.New("Get")
}

func (_ *testStorage) GetAll() ([]*models.Todo, error) {
	return nil, errors.New("GetAll")
}

func (_ *testStorage) FinishTodo(models.ID) error {
	return errors.New("Finish")
}

func Test_App(t *testing.T) {
	db := models.NewInMemoryStorage()
	a := New(db)
	require.NotNil(t, a)

	a.initRoutes()

	req := httptest.NewRequest("POST", "/todo/create", strings.NewReader("{}"))
	w := httptest.NewRecorder()
	a.router.ServeHTTP(w, req)
	require.Equal(t, w.Code, http.StatusBadRequest)
	require.Equal(t, "title is required", w.Body.String())

	req = httptest.NewRequest(
		"POST",
		"/todo/create",
		strings.NewReader(`{"title":"A","content":"a"}`),
	)
	w = httptest.NewRecorder()
	a.router.ServeHTTP(w, req)
	require.Equal(t, w.Code, http.StatusCreated)

	req = httptest.NewRequest("GET", "/todo/0", strings.NewReader(""))
	w = httptest.NewRecorder()
	a.router.ServeHTTP(w, req)
	require.Equal(t, w.Code, http.StatusOK)
	require.Equal(t, w.Header().Get("Content-Type"), "application/json")
	require.Equal(t, w.Body.String(), `{"id":0,"title":"A","content":"a","finished":false}`)

	req = httptest.NewRequest("POST", "/todo/0/finish", strings.NewReader(``))
	w = httptest.NewRecorder()
	a.router.ServeHTTP(w, req)
	require.Equal(t, w.Code, http.StatusOK)

	req = httptest.NewRequest("GET", "/todo", strings.NewReader(``))
	w = httptest.NewRecorder()
	a.router.ServeHTTP(w, req)
	require.Equal(t, w.Code, http.StatusOK)

	req = httptest.NewRequest("GET", "/", strings.NewReader(``))
	w = httptest.NewRecorder()
	a.router.ServeHTTP(w, req)
	require.Equal(t, w.Code, http.StatusOK)
	require.Equal(t, w.Body.String(), `"API is up and working!"`)

	req = httptest.NewRequest("POST", "/todo/7/finish", strings.NewReader(``))
	w = httptest.NewRecorder()
	a.router.ServeHTTP(w, req)
	require.Equal(t, w.Code, http.StatusInternalServerError)

	req = httptest.NewRequest("GET", "/todo/9", strings.NewReader(""))
	w = httptest.NewRecorder()
	a.router.ServeHTTP(w, req)
	require.Equal(t, w.Code, http.StatusInternalServerError)

	req = httptest.NewRequest("POST", "/todo/create", strings.NewReader("f"))
	w = httptest.NewRecorder()
	a.router.ServeHTTP(w, req)
	require.Equal(t, w.Code, http.StatusBadRequest)
	require.Equal(t, "payload is required", w.Body.String())

	id := strings.Repeat("9", 30)
	req = httptest.NewRequest("GET", "/todo/"+id, strings.NewReader(""))
	w = httptest.NewRecorder()
	a.router.ServeHTTP(w, req)
	require.Equal(t, w.Code, http.StatusBadRequest)

	req = httptest.NewRequest("POST", "/todo/"+id+"/finish", strings.NewReader(``))
	w = httptest.NewRecorder()
	a.router.ServeHTTP(w, req)
	require.Equal(t, w.Code, http.StatusBadRequest)
}

func Test_FakeDb(t *testing.T) {
	db := &testStorage{}
	a := New(db)
	require.NotNil(t, a)

	a.initRoutes()

	req := httptest.NewRequest(
		"POST",
		"/todo/create",
		strings.NewReader(`{"title":"A","content":"a"}`),
	)
	w := httptest.NewRecorder()
	a.router.ServeHTTP(w, req)
	require.Equal(t, w.Code, http.StatusInternalServerError)

	req = httptest.NewRequest("GET", "/todo", strings.NewReader(`{"title":"A","content":"a"}`))
	w = httptest.NewRecorder()
	a.router.ServeHTTP(w, req)
	require.Equal(t, w.Code, http.StatusInternalServerError)

	req = httptest.NewRequest(
		"POST",
		"/todo/1/finish",
		strings.NewReader(`{"title":"A","content":"a"}`),
	)
	w = httptest.NewRecorder()
	a.router.ServeHTTP(w, req)
	require.Equal(t, w.Code, http.StatusInternalServerError)

	req = httptest.NewRequest("GET", "/todo/0", strings.NewReader(`{"title":"A","content":"a"}`))
	w = httptest.NewRecorder()
	a.router.ServeHTTP(w, req)
	require.Equal(t, w.Code, http.StatusInternalServerError)
}
