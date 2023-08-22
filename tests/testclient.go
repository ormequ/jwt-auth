package tests

import (
	"bytes"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/goccy/go-json"
	"go.mongodb.org/mongo-driver/mongo"
	"io"
	"jwt-auth/internal/adapters/bcrypt"
	repo "jwt-auth/internal/adapters/mongo"
	"jwt-auth/internal/app"
	"jwt-auth/internal/httpserver"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"time"
)

var (
	ErrBadRequest = fmt.Errorf("bad request")
	ErrForbidden  = fmt.Errorf("forbidden")
	ErrNotFound   = fmt.Errorf("not found")
)

const (
	accessSecret  = "access-test-secret"
	refreshSecret = "refresh-test-secret"
)

var db *mongo.Client

func setupClient(accessExp time.Duration, refreshExp time.Duration) *testClient {
	a := app.New(
		repo.New(db.Database("test")),
		bcrypt.New(10),
		accessSecret,
		refreshSecret,
		accessExp,
		refreshExp,
	)
	srv := httpserver.New(slog.Default(), ":18080", gin.ReleaseMode, a)
	testSrv := httptest.NewServer(srv.Handler)

	return &testClient{
		client:  testSrv.Client(),
		baseURL: testSrv.URL,
	}
}

type testClient struct {
	client  *http.Client
	baseURL string
}

func (tc *testClient) request(body map[string]any, method string, endpoint string, out any) error {
	data, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("unable to marshal: %w", err)
	}

	req, err := http.NewRequest(method, tc.baseURL+"/api/"+endpoint, bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("unable to create request: %w", err)
	}
	req.Header.Add("Content-Type", "application/json")
	resp, err := tc.client.Do(req)
	if err != nil {
		return fmt.Errorf("unexpected error: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		if resp.StatusCode == http.StatusNotFound {
			return ErrNotFound
		}
		if resp.StatusCode == http.StatusBadRequest {
			return ErrBadRequest
		}
		if resp.StatusCode == http.StatusForbidden {
			return ErrForbidden
		}
		return fmt.Errorf("unexpected status code: %s", resp.Status)
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("unable to read response: %w", err)
	}

	err = json.Unmarshal(respBody, out)
	if err != nil {
		return fmt.Errorf("unable to unmarshal: %w", err)
	}

	return nil
}

type jwtPair httpserver.JWTPairResponse

type jwtPairResponse struct {
	Data jwtPair `json:"data"`
}

func (tc *testClient) generate(userID string) (jwtPair, error) {
	body := map[string]any{
		"user_id": userID,
	}
	var response jwtPairResponse
	err := tc.request(body, http.MethodPost, "generate", &response)
	return response.Data, err
}

func (tc *testClient) refresh(token string) (jwtPair, error) {
	body := map[string]any{
		"token": token,
	}
	var response jwtPairResponse
	err := tc.request(body, http.MethodPut, "refresh", &response)
	return response.Data, err
}
