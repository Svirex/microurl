package handlers

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/Svirex/microurl/internal/generators"
	"github.com/Svirex/microurl/internal/pkg/config"
	"github.com/Svirex/microurl/internal/pkg/context"
	"github.com/Svirex/microurl/internal/pkg/repositories"
	"github.com/Svirex/microurl/internal/storage"
	"github.com/stretchr/testify/assert"
)

// Test 1: Пустое тело запроса
// Test 2: Подменить репозиторий, который будет выдавать ошибку на попытку добавления
// Test 3: Передать правильные параметры, проверить ответ регуляркой

type MockRepository struct{}

var _ repositories.Repository = &MockRepository{}

func (m *MockRepository) Add(shortID, url string) error {
	return fmt.Errorf("couldn't add")
}

func (m *MockRepository) Get(shortID string) (*string, error) {
	result := "res"
	return &result, nil
}

func TestPost(t *testing.T) {
	config := config.Config{
		Host: "localhost",
		Port: 8080,
	}
	appCtx := &context.AppContext{
		Config:     &config,
		Generator:  generators.NewSimpleGenerator(time.Now().UnixNano()),
		Repository: storage.NewMapRepository(),
	}
	t.Run("empty body", func(t *testing.T) {
		r := httptest.NewRequest(http.MethodPost, "/", nil)
		w := httptest.NewRecorder()

		Post(appCtx)(w, r)

		res := w.Result()

		assert.Equal(t, http.StatusBadRequest, res.StatusCode)
		defer res.Body.Close()
		resBody, err := io.ReadAll(res.Body)

		assert.NoError(t, err)

		assert.Equal(t, "", string(resBody))
	})
	t.Run("good", func(t *testing.T) {
		body := strings.NewReader("http://svirex.ru")
		r := httptest.NewRequest(http.MethodPost, "/", body)
		w := httptest.NewRecorder()

		Post(appCtx)(w, r)

		res := w.Result()

		assert.Equal(t, http.StatusCreated, res.StatusCode)
		defer res.Body.Close()
		resBody, err := io.ReadAll(res.Body)

		assert.NoError(t, err)

		reg := regexp.MustCompile(fmt.Sprintf("^http://%s:%d/[A-Za-z]+$", config.Host, config.Port))
		assert.True(t, reg.MatchString(string(resBody)))
	})
	t.Run("couldn't add", func(t *testing.T) {
		appCtx := &context.AppContext{
			Config:     &config,
			Generator:  generators.NewSimpleGenerator(time.Now().UnixNano()),
			Repository: &MockRepository{},
		}
		body := strings.NewReader("http://svirex.ru")
		r := httptest.NewRequest(http.MethodPost, "/", body)
		w := httptest.NewRecorder()

		Post(appCtx)(w, r)

		res := w.Result()
		assert.Equal(t, http.StatusBadRequest, res.StatusCode)
		defer res.Body.Close()
		resBody, err := io.ReadAll(res.Body)

		assert.NoError(t, err)

		assert.Equal(t, "", string(resBody))
	})
}
