package server

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"regexp"
	"strings"
	"testing"

	"github.com/Svirex/microurl/internal/apis"
	"github.com/Svirex/microurl/internal/generators"
	"github.com/Svirex/microurl/internal/pkg/logging"
	"github.com/Svirex/microurl/internal/pkg/models"
	"github.com/Svirex/microurl/internal/pkg/repositories"
	"github.com/Svirex/microurl/internal/services"
	"github.com/Svirex/microurl/internal/storage"
	"github.com/stretchr/testify/require"
)

type FakeLogger struct{}

var _ logging.Logger = (*FakeLogger)(nil)

func (*FakeLogger) Info(params ...any)  {}
func (*FakeLogger) Error(params ...any) {}
func (*FakeLogger) Shutdown() error     { return nil }

func TestRouterPost(t *testing.T) {
	rep := storage.NewMapRepository()
	require.NotNil(t, rep)
	service := services.NewShortenerService(generators.NewSimpleGenerator(255), rep, 8)
	require.NotNil(t, service)
	api := apis.NewShortenerAPI(service, "http://svirex.ru")
	require.NotNil(t, api)

	testServer := httptest.NewServer(api.Routes(&FakeLogger{}))
	defer testServer.Close()

	u, _ := url.Parse(testServer.URL)
	api.BaseURL = "http://" + u.Host

	{
		req, err := http.NewRequest(http.MethodPost, testServer.URL, nil)
		require.NoError(t, err)

		resp, err := testServer.Client().Do(req)
		require.NoError(t, err)
		resp.Body.Close()

		require.Equal(t, http.StatusBadRequest, resp.StatusCode)
	}

	{
		req, err := http.NewRequest(http.MethodPost, testServer.URL, strings.NewReader("http://svirex.ru"))
		require.NoError(t, err)

		resp, err := testServer.Client().Do(req)
		require.NoError(t, err)

		body, err := io.ReadAll(resp.Body)
		require.NoError(t, err)
		resp.Body.Close()

		require.Equal(t, http.StatusCreated, resp.StatusCode)

		reg := regexp.MustCompile(fmt.Sprintf("^%s/[A-Za-z]+$", testServer.URL))
		require.True(t, reg.MatchString(string(body)))

	}

}

type MockRepository struct{}

var _ repositories.URLRepository = (*MockRepository)(nil)

func (m *MockRepository) Add(context.Context, *models.RepositoryAddRecord) error {
	return fmt.Errorf("couldn't add")
}

func (m *MockRepository) Get(context.Context, *models.RepositoryGetRecord) (*models.RepositoryGetResult, error) {
	return models.NewRepositoryGetResult("res"), nil
}

func (m *MockRepository) Shutdown() error {
	return nil
}

func TestRouterPostWithMockRepo(t *testing.T) {
	service := services.NewShortenerService(generators.NewSimpleGenerator(255), &MockRepository{}, 8)
	require.NotNil(t, service)
	api := apis.NewShortenerAPI(service, "http://svirex.ru")
	require.NotNil(t, api)

	testServer := httptest.NewServer(api.Routes(&FakeLogger{}))
	defer testServer.Close()

	{
		req, err := http.NewRequest(http.MethodPost, testServer.URL, strings.NewReader("http://svirex.ru"))
		require.NoError(t, err)

		resp, err := testServer.Client().Do(req)
		require.NoError(t, err)
		resp.Body.Close()

		require.Equal(t, http.StatusBadRequest, resp.StatusCode)

	}

}

func TestServerGet(t *testing.T) {
	rep := storage.NewMapRepository()
	require.NotNil(t, rep)
	service := services.NewShortenerService(generators.NewSimpleGenerator(255), rep, 8)
	require.NotNil(t, service)
	api := apis.NewShortenerAPI(service, "http://svirex.ru")
	require.NotNil(t, api)

	testServer := httptest.NewServer(api.Routes(&FakeLogger{}))
	defer testServer.Close()

	{
		req, err := http.NewRequest(http.MethodGet, testServer.URL, nil)
		require.NoError(t, err)

		resp, err := testServer.Client().Do(req)
		require.NoError(t, err)
		resp.Body.Close()

		require.Equal(t, http.StatusMethodNotAllowed, resp.StatusCode)
	}
	{
		req, err := http.NewRequest(http.MethodGet, testServer.URL+"/Egege/DDFSDE", nil)
		require.NoError(t, err)

		resp, err := testServer.Client().Do(req)
		require.NoError(t, err)
		resp.Body.Close()

		require.Equal(t, http.StatusNotFound, resp.StatusCode)
	}
	{
		req, err := http.NewRequest(http.MethodGet, testServer.URL+"/EgegeY", nil)
		require.NoError(t, err)

		resp, err := testServer.Client().Do(req)
		require.NoError(t, err)
		resp.Body.Close()

		require.Equal(t, http.StatusBadRequest, resp.StatusCode)
	}
	{
		sourceURL := "http://svirex.ru"
		req, err := http.NewRequest(http.MethodPost, testServer.URL, strings.NewReader(sourceURL))
		require.NoError(t, err)

		resp, err := testServer.Client().Do(req)
		require.NoError(t, err)

		body, err := io.ReadAll(resp.Body)
		require.NoError(t, err)
		resp.Body.Close()

		url := string(body)

		splitted := strings.Split(url, "/")
		shortID := splitted[len(splitted)-1]
		req, err = http.NewRequest(http.MethodGet, fmt.Sprintf("%s/%s", testServer.URL, shortID), nil)
		require.NoError(t, err)

		client := testServer.Client()
		client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		}

		resp, err = client.Do(req)
		require.NoError(t, err)

		url = resp.Header.Get("Location")

		require.Equal(t, sourceURL, url)

		require.Equal(t, http.StatusTemporaryRedirect, resp.StatusCode)

		resp.Body.Close()

	}
}

func TestServerJSONShorten(t *testing.T) {
	rep := storage.NewMapRepository()
	require.NotNil(t, rep)
	service := services.NewShortenerService(generators.NewSimpleGenerator(255), rep, 8)
	require.NotNil(t, service)
	api := apis.NewShortenerAPI(service, "http://svirex.ru")
	require.NotNil(t, api)

	testServer := httptest.NewServer(api.Routes(&FakeLogger{}))
	defer testServer.Close()

	apiURL := testServer.URL + "/api/shorten"
	// Not exists application/json header
	{
		req, err := http.NewRequest(http.MethodPost, apiURL, http.NoBody)
		require.NoError(t, err)

		resp, err := testServer.Client().Do(req)
		require.NoError(t, err)
		resp.Body.Close()

		require.Equal(t, http.StatusBadRequest, resp.StatusCode)
	}
	// Empty body
	{
		req, err := http.NewRequest(http.MethodPost, apiURL, http.NoBody)
		require.NoError(t, err)

		req.Header.Set("Content-Type", "application/json")

		resp, err := testServer.Client().Do(req)
		require.NoError(t, err)
		resp.Body.Close()

		require.Equal(t, http.StatusBadRequest, resp.StatusCode)
	}
	// Not JSON body
	{
		req, err := http.NewRequest(http.MethodPost, apiURL, strings.NewReader("http://svirex.ru"))
		require.NoError(t, err)

		req.Header.Set("Content-Type", "application/json")

		resp, err := testServer.Client().Do(req)
		require.NoError(t, err)
		resp.Body.Close()

		require.Equal(t, http.StatusBadRequest, resp.StatusCode)
	}
	// Good
	{
		in := models.InputJSON{
			URL: "http://svirex.ru",
		}
		body, _ := json.Marshal(in)
		req, err := http.NewRequest(http.MethodPost, apiURL, bytes.NewReader(body))
		require.NoError(t, err)

		req.Header.Set("Content-Type", "application/json")

		resp, err := testServer.Client().Do(req)
		require.NoError(t, err)

		require.Equal(t, http.StatusCreated, resp.StatusCode)
		require.Equal(t, resp.Header.Get("Content-Type"), "application/json")

		body, err = io.ReadAll(resp.Body)
		resp.Body.Close()
		require.NoError(t, err)

		var resultJSON models.ResultJSON
		err = json.Unmarshal(body, &resultJSON)
		require.NoError(t, err)

		reg := regexp.MustCompile(fmt.Sprintf("^%s/[A-Za-z]+$", "http://svirex.ru"))
		fmt.Println(fmt.Sprintf("^%s/[A-Za-z]+$", "http://svirex.ru"), resultJSON.ShortURL)
		require.True(t, reg.MatchString(resultJSON.ShortURL))
	}

}
