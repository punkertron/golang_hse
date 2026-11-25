package models_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"gitlab.com/slon/shad-go/10-Generics/coverme/models"
)

func Test_Models(t *testing.T) {
	s := models.NewInMemoryStorage()
	require.NotNil(t, s, "model is nil")

	id, _ := s.AddTodo("Key", "Value")
	todo, _ := s.GetTodo(id.ID)
	require.Equal(t, todo.Title, "Key", "incorrect key")

	require.NoError(t, s.FinishTodo(id.ID))

	tasks, _ := s.GetAll()
	require.Equal(t, len(tasks), 1)

	_, err := s.GetTodo(id.ID + 1)
	require.Error(t, err)
}
