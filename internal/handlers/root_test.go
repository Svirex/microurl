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

		Post(w, r, appCtx)

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

		Post(w, r, appCtx)

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

		Post(w, r, appCtx)

		res := w.Result()
		assert.Equal(t, http.StatusBadRequest, res.StatusCode)
		defer res.Body.Close()
		resBody, err := io.ReadAll(res.Body)

		assert.NoError(t, err)

		assert.Equal(t, "", string(resBody))
	})
}

// Test 1: invalid url
// Test 2: not found in repository
// Test 3: good request, post before
func TestGet(t *testing.T) {
	config := config.Config{
		Host: "localhost",
		Port: 8080,
	}
	appCtx := &context.AppContext{
		Config:     &config,
		Generator:  generators.NewSimpleGenerator(time.Now().UnixNano()),
		Repository: storage.NewMapRepository(),
	}
	t.Run("invalid url #1", func(t *testing.T) {
		r := httptest.NewRequest(http.MethodGet, "/ETGFDT/DDFDF", nil)
		w := httptest.NewRecorder()

		Get(w, r, appCtx)

		res := w.Result()
		assert.Equal(t, http.StatusBadRequest, res.StatusCode)
		defer res.Body.Close()
		resBody, err := io.ReadAll(res.Body)

		assert.NoError(t, err)

		assert.Equal(t, "", string(resBody))

	})
	t.Run("invalid url #2", func(t *testing.T) {
		appCtx := &context.AppContext{
			Config:     &config,
			Generator:  generators.NewSimpleGenerator(time.Now().UnixNano()),
			Repository: storage.NewMapRepository(),
		}
		appCtx.Repository.Add("", "http://svirex.ru")
		r := httptest.NewRequest(http.MethodGet, "/", nil)
		w := httptest.NewRecorder()

		Get(w, r, appCtx)

		res := w.Result()
		assert.Equal(t, http.StatusBadRequest, res.StatusCode)
		defer res.Body.Close()
		resBody, err := io.ReadAll(res.Body)

		assert.NoError(t, err)

		assert.Equal(t, "", string(resBody))

	})
	t.Run("not found", func(t *testing.T) {
		r := httptest.NewRequest(http.MethodGet, "/EAftGY", nil)
		w := httptest.NewRecorder()

		Get(w, r, appCtx)

		res := w.Result()
		assert.Equal(t, http.StatusBadRequest, res.StatusCode)
		defer res.Body.Close()
		resBody, err := io.ReadAll(res.Body)

		assert.NoError(t, err)

		assert.Equal(t, "", string(resBody))

	})
	t.Run("good", func(t *testing.T) {
		sourceURL := "http://svirex.ru"
		body := strings.NewReader(sourceURL)
		postR := httptest.NewRequest(http.MethodPost, "/", body)
		postW := httptest.NewRecorder()

		Post(postW, postR, appCtx)

		postRes := postW.Result()
		defer postRes.Body.Close()
		resBody, err := io.ReadAll(postRes.Body)

		assert.NoError(t, err)

		url := string(resBody)

		splitted := strings.Split(url, "/")
		shortID := splitted[len(splitted)-1]

		r := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/%s", shortID), nil)
		w := httptest.NewRecorder()

		Get(w, r, appCtx)

		res := w.Result()
		assert.Equal(t, http.StatusTemporaryRedirect, res.StatusCode)

		url = res.Header.Get("Location")
		assert.Equal(t, sourceURL, url)
		defer res.Body.Close()
	})
}
