package server

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"regexp"
	"strings"
	"testing"

	"github.com/Svirex/microurl/internal/generators"
	"github.com/Svirex/microurl/internal/pkg/config"
	"github.com/Svirex/microurl/internal/pkg/context"
	"github.com/Svirex/microurl/internal/pkg/repositories"
	"github.com/Svirex/microurl/internal/storage"
	"github.com/stretchr/testify/require"
)

func TestRouterPost(t *testing.T) {
	config := &config.Config{}
	appCtx := &context.AppContext{
		Config:     config,
		Repository: storage.NewMapRepository(),
		Generator:  generators.NewSimpleGenerator(255),
	}

	testServer := httptest.NewServer(MainRoutes(appCtx))
	defer testServer.Close()
	u, _ := url.Parse(testServer.URL)
	appCtx.Config.Addr = u.Host
	appCtx.Config.BaseURL = "http://" + u.Host

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
		fmt.Println(fmt.Sprintf("^%s/[A-Za-z]+$", testServer.URL), string(body))
		require.True(t, reg.MatchString(string(body)))

	}

}

type MockRepository struct{}

var _ repositories.Repository = &MockRepository{}

func (m *MockRepository) Add(shortID, url string) error {
	return fmt.Errorf("couldn't add")
}

func (m *MockRepository) Get(shortID string) (*string, error) {
	result := "res"
	return &result, nil
}

func TestRouterPostWithMockRepo(t *testing.T) {
	config := &config.Config{}
	appCtx := &context.AppContext{
		Config:     config,
		Repository: &MockRepository{},
		Generator:  generators.NewSimpleGenerator(255),
	}

	testServer := httptest.NewServer(MainRoutes(appCtx))
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
	config := &config.Config{}
	appCtx := &context.AppContext{
		Config:     config,
		Repository: storage.NewMapRepository(),
		Generator:  generators.NewSimpleGenerator(255),
	}

	testServer := httptest.NewServer(MainRoutes(appCtx))
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
