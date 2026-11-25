package client_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.com/slon/shad-go/10-Generics/coverme/client"
	"gitlab.com/slon/shad-go/10-Generics/coverme/models"
)

func Test_Get(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Equal(t, "/todo/7", r.URL.Path)

		_, _ = io.WriteString(w, `{"id":7,"title":"Read","finished":false}`)
	}))
	defer server.Close()

	todo, err := client.New(server.URL).Get(7)
	require.NoError(t, err)
	require.NotNil(t, todo)
	require.Equal(t, models.ID(7), todo.ID)
	require.Equal(t, "Read", todo.Title)
}

func Test_Add(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "/todo/create", r.URL.Path)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

		body, err := io.ReadAll(r.Body)
		if assert.NoError(t, err) {
			assert.JSONEq(t, `{"title":"A","content":"a"}`, string(body))
		}

		w.WriteHeader(http.StatusCreated)
		_, _ = io.WriteString(w, `{"id":7,"title":"A","content":"a"}`)
	}))
	defer server.Close()

	req := &models.AddRequest{
		Title:   "A",
		Content: "a",
	}
	todo, err := client.New(server.URL).Add(req)
	require.NoError(t, err)
	require.NotNil(t, todo)
	require.Equal(t, models.ID(7), todo.ID)
	require.Equal(t, "A", todo.Title)
	require.Equal(t, "a", todo.Content)
}

func Test_List(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Equal(t, "/todo", r.URL.Path)

		_, _ = io.WriteString(
			w,
			`[{"id":7,"title":"Read"},{"id":8,"title":"Write","finished":true}]`,
		)
	}))
	defer server.Close()

	todos, err := client.New(server.URL).List()
	require.NoError(t, err)
	require.Equal(t, []*models.Todo{
		{ID: 7, Title: "Read"},
		{ID: 8, Title: "Write", Finished: true},
	}, todos)
}

func Test_Finish(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "/todo/7/finish", r.URL.Path)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	require.NoError(t, client.New(server.URL).Finish(7))
}

func Test_ClientErrors(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer server.Close()

	calls := []struct {
		name string
		call func(*client.Client) error
	}{
		{"Add", func(c *client.Client) error {
			_, err := c.Add(&models.AddRequest{Title: "Read"})
			return err
		}},
		{"Get", func(c *client.Client) error {
			_, err := c.Get(7)
			return err
		}},
		{"List", func(c *client.Client) error {
			_, err := c.List()
			return err
		}},
		{"Finish", func(c *client.Client) error {
			return c.Finish(7)
		}},
	}

	for _, tc := range calls {
		t.Run(tc.name+"/unexpected status", func(t *testing.T) {
			require.ErrorContains(t, tc.call(client.New(server.URL)), "unexpected status code 503")
		})
		t.Run(tc.name+"/invalid address", func(t *testing.T) {
			require.Error(t, tc.call(client.New("://invalid")))
		})
	}
}
