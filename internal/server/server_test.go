package server

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
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
	"github.com/Svirex/microurl/internal/logging"
	"github.com/Svirex/microurl/internal/models"
	"github.com/Svirex/microurl/internal/services"
	"github.com/Svirex/microurl/internal/storage"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/require"
)

type FakeLogger struct{}

var _ logging.Logger = (*FakeLogger)(nil)

func (*FakeLogger) Info(params ...any)  {}
func (*FakeLogger) Error(params ...any) {}
func (*FakeLogger) Shutdown() error     { return nil }

type MockDBCheck struct {
	err error
}

var _ services.DBCheck = (*MockDBCheck)(nil)

func (m *MockDBCheck) Ping(ctx context.Context) error {
	return m.err
}

func (m *MockDBCheck) Shutdown() error {
	return nil
}

func NewMockDBCheck(err error) *MockDBCheck {
	return &MockDBCheck{
		err: err,
	}
}

func TestRouterPost(t *testing.T) {
	rep := storage.NewMapRepository()
	require.NotNil(t, rep)
	service := services.NewShortenerService(generators.NewSimpleGenerator(255), rep, 8)
	require.NotNil(t, service)
	api := apis.NewShortenerAPI(service, NewMockDBCheck(nil), "http://svirex.ru")
	require.NotNil(t, api)

	testServer := httptest.NewServer(api.Routes(&FakeLogger{}, "fake_secret_key"))
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

var _ storage.URLRepository = (*MockRepository)(nil)

func (m *MockRepository) Add(context.Context, *models.RepositoryAddRecord) (*models.RepositoryGetRecord, error) {
	return nil, fmt.Errorf("couldn't add")
}

func (m *MockRepository) Get(context.Context, *models.RepositoryGetRecord) (*models.RepositoryGetResult, error) {
	return models.NewRepositoryGetResult("res"), nil
}

func (m *MockRepository) Shutdown() error {
	return nil
}

func (m *MockRepository) Batch(context.Context, *models.BatchService) (*models.BatchResponse, error) {
	return nil, fmt.Errorf("couldn't add")
}

func (m *MockRepository) UserURLs(_ context.Context, uid string) ([]models.UserURLRecord, error) {
	result := make([]models.UserURLRecord, 0)

	return result, nil
}

func TestRouterPostWithMockRepo(t *testing.T) {
	service := services.NewShortenerService(generators.NewSimpleGenerator(255), &MockRepository{}, 8)
	require.NotNil(t, service)
	api := apis.NewShortenerAPI(service, NewMockDBCheck(nil), "http://svirex.ru")
	require.NotNil(t, api)

	testServer := httptest.NewServer(api.Routes(&FakeLogger{}, "fake_secret_key"))
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
	api := apis.NewShortenerAPI(service, NewMockDBCheck(nil), "http://svirex.ru")
	require.NotNil(t, api)

	testServer := httptest.NewServer(api.Routes(&FakeLogger{}, "fake_secret_key"))
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
	api := apis.NewShortenerAPI(service, NewMockDBCheck(nil), "http://svirex.ru")
	require.NotNil(t, api)

	testServer := httptest.NewServer(api.Routes(&FakeLogger{}, "fake_secret_key"))
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
		require.Equal(t, "application/json", resp.Header.Get("Content-Type"))

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

func TestServerPingOk(t *testing.T) {
	rep := storage.NewMapRepository()
	require.NotNil(t, rep)
	service := services.NewShortenerService(generators.NewSimpleGenerator(255), rep, 8)
	require.NotNil(t, service)
	api := apis.NewShortenerAPI(service, NewMockDBCheck(nil), "http://svirex.ru")
	require.NotNil(t, api)

	testServer := httptest.NewServer(api.Routes(&FakeLogger{}, "fake_secret_key"))
	defer testServer.Close()

	{
		req, err := http.NewRequest(http.MethodGet, testServer.URL+"/ping", nil)
		require.NoError(t, err)

		resp, err := testServer.Client().Do(req)
		require.NoError(t, err)
		resp.Body.Close()

		require.Equal(t, http.StatusOK, resp.StatusCode)
	}
}

func TestServerPingFail(t *testing.T) {
	rep := storage.NewMapRepository()
	require.NotNil(t, rep)
	service := services.NewShortenerService(generators.NewSimpleGenerator(255), rep, 8)
	require.NotNil(t, service)
	api := apis.NewShortenerAPI(service, NewMockDBCheck(errors.New("OOPS!!!")), "http://svirex.ru")
	require.NotNil(t, api)

	testServer := httptest.NewServer(api.Routes(&FakeLogger{}, "fake_secret_key"))
	defer testServer.Close()

	{
		req, err := http.NewRequest(http.MethodGet, testServer.URL+"/ping", nil)
		require.NoError(t, err)

		resp, err := testServer.Client().Do(req)
		require.NoError(t, err)
		resp.Body.Close()

		require.Equal(t, http.StatusInternalServerError, resp.StatusCode)
	}
}

func TestBatch(t *testing.T) {
	rep := storage.NewMapRepository()
	require.NotNil(t, rep)
	service := services.NewShortenerService(generators.NewSimpleGenerator(255), rep, 8)
	require.NotNil(t, service)
	api := apis.NewShortenerAPI(service, NewMockDBCheck(errors.New("OOPS!!!")), "http://svirex.ru")
	require.NotNil(t, api)

	testServer := httptest.NewServer(api.Routes(&FakeLogger{}, "fake_secret_key"))
	defer testServer.Close()

	{
		body := []models.BatchRequestRecord{
			{
				CorrID: "1",
				URL:    "https://ya.ru",
			},
			{
				CorrID: "2",
				URL:    "https://ya.ru",
			},
		}
		bodyBytes, err := json.Marshal(body)
		require.NoError(t, err)

		req, err := http.NewRequest(http.MethodPost, testServer.URL+"/api/shorten/batch", bytes.NewReader(bodyBytes))
		require.NoError(t, err)

		req.Header.Set("Content-Type", "application/json")

		resp, err := testServer.Client().Do(req)
		require.NoError(t, err)

		respBody, err := io.ReadAll(resp.Body)
		require.NoError(t, err)
		resp.Body.Close()

		require.Equal(t, http.StatusCreated, resp.StatusCode)

		require.NotEmpty(t, respBody)
	}
}

func NewTestServerWithMapRepository(t *testing.T) *httptest.Server {
	rep := storage.NewMapRepository()
	require.NotNil(t, rep)
	service := services.NewShortenerService(generators.NewSimpleGenerator(255), rep, 8)
	require.NotNil(t, service)
	api := apis.NewShortenerAPI(service, NewMockDBCheck(nil), "http://svirex.ru")
	require.NotNil(t, api)

	secretKey := "fake_secret_key"

	return httptest.NewServer(api.Routes(&FakeLogger{}, secretKey))
}

func TestSetupAuthCookie(t *testing.T) {
	rep := storage.NewMapRepository()
	require.NotNil(t, rep)
	service := services.NewShortenerService(generators.NewSimpleGenerator(255), rep, 8)
	require.NotNil(t, service)
	api := apis.NewShortenerAPI(service, NewMockDBCheck(nil), "http://svirex.ru")
	require.NotNil(t, api)

	secretKey := "fake_secret_key"

	testServer := httptest.NewServer(api.Routes(&FakeLogger{}, secretKey))
	defer testServer.Close()

	{ // Нет куки
		req, err := http.NewRequest(http.MethodGet, testServer.URL+"/api/user/urls", nil)
		require.NoError(t, err)

		resp, err := testServer.Client().Do(req)
		require.NoError(t, err)
		resp.Body.Close()

		foundJWTCookie := false
		for _, v := range resp.Cookies() {
			if v.Name == "jwt" {
				foundJWTCookie = true
				break
			}
		}
		require.True(t, foundJWTCookie)
	}
	{ // Невалидный токен jwt
		req, err := http.NewRequest(http.MethodGet, testServer.URL+"/api/user/urls", nil)
		require.NoError(t, err)

		req.AddCookie(&http.Cookie{
			Name:  "jwt",
			Value: "sfgdfsdfs",
		})

		resp, err := testServer.Client().Do(req)
		require.NoError(t, err)
		resp.Body.Close()

		foundJWTCookie := false
		for _, v := range resp.Cookies() {
			if v.Name == "jwt" {
				foundJWTCookie = true
				break
			}
		}
		require.True(t, foundJWTCookie)
	}
	{ // В jwt нет user id
		req, err := http.NewRequest(http.MethodGet, testServer.URL+"/api/user/urls", nil)
		require.NoError(t, err)

		token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{})
		tokenString, err := token.SignedString([]byte(secretKey))
		require.NoError(t, err)

		req.AddCookie(&http.Cookie{
			Name:  "jwt",
			Value: tokenString,
		})

		resp, err := testServer.Client().Do(req)
		require.NoError(t, err)
		resp.Body.Close()

		require.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	}
}

func TestUserURLsWithMapRepository(t *testing.T) {
	testServer := NewTestServerWithMapRepository()
	defer testServer.Close()
}
