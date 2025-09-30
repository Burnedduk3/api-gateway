package handlers_test

import (
	"api-gateway/pkg/logger"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"api-gateway/internal/adapters/http/handlers"
	"api-gateway/internal/application/dto"

	"github.com/stretchr/testify/assert"
)

func TestProxyClient_Forward_Success(t *testing.T) {
	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/users", r.URL.Path)
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"message":"success"}`))
	}))
	defer server.Close()
	log := logger.New("test")

	client := handlers.NewProxyClient(log, 30*time.Second)

	request := &dto.ProxyRequest{
		URL:    server.URL + "/api/users",
		Method: "GET",
		Headers: map[string][]string{
			"Content-Type": {"application/json"},
		},
		Body: []byte{},
	}

	response, err := client.Forward(context.Background(), request)

	assert.NoError(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, http.StatusOK, response.StatusCode)
	assert.Equal(t, "application/json", response.Headers.Get("Content-Type"))
	assert.Contains(t, string(response.Body), "success")
}

func TestProxyClient_Forward_WithBody(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)

		// Read and verify body
		buf := make([]byte, 1024)
		n, _ := r.Body.Read(buf)
		assert.Equal(t, `{"name":"John"}`, string(buf[:n]))

		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(`{"id":1,"name":"John"}`))
	}))
	defer server.Close()
	log := logger.New("test")

	client := handlers.NewProxyClient(log, 30*time.Second)

	request := &dto.ProxyRequest{
		URL:    server.URL + "/api/users",
		Method: "POST",
		Headers: map[string][]string{
			"Content-Type": {"application/json"},
		},
		Body: []byte(`{"name":"John"}`),
	}

	response, err := client.Forward(context.Background(), request)

	assert.NoError(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, http.StatusCreated, response.StatusCode)
	assert.Contains(t, string(response.Body), "John")
}

func TestProxyClient_Forward_Timeout(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Simulate slow response
		time.Sleep(2 * time.Second)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()
	log := logger.New("test")

	// Create client with short timeout
	client := handlers.NewProxyClient(log, 30*time.Second)

	request := &dto.ProxyRequest{
		URL:    server.URL + "/api/slow",
		Method: "GET",
	}

	response, err := client.Forward(context.Background(), request)

	assert.Error(t, err)
	assert.Nil(t, response)
	assert.Contains(t, err.Error(), "timeout")
}

func TestProxyClient_Forward_InvalidURL(t *testing.T) {
	log := logger.New("test")
	client := handlers.NewProxyClient(log, 30*time.Second)

	request := &dto.ProxyRequest{
		URL:    "http://invalid-host-that-does-not-exist:9999/api/users",
		Method: "GET",
	}

	response, err := client.Forward(context.Background(), request)

	assert.Error(t, err)
	assert.Nil(t, response)
}

func TestProxyClient_Forward_ServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error":"internal server error"}`))
	}))
	log := logger.New("test")

	defer server.Close()

	client := handlers.NewProxyClient(log, 30*time.Second)

	request := &dto.ProxyRequest{
		URL:    server.URL + "/api/users",
		Method: "GET",
	}

	response, err := client.Forward(context.Background(), request)

	assert.NoError(t, err) // Connection succeeded, just returned error status
	assert.NotNil(t, response)
	assert.Equal(t, http.StatusInternalServerError, response.StatusCode)
}

func TestProxyClient_Forward_ContextCancellation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(1 * time.Second)
		w.WriteHeader(http.StatusOK)
	}))
	log := logger.New("test")

	defer server.Close()

	client := handlers.NewProxyClient(log, 30*time.Second)

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	request := &dto.ProxyRequest{
		URL:    server.URL + "/api/users",
		Method: "GET",
	}

	response, err := client.Forward(ctx, request)

	assert.Error(t, err)
	assert.Nil(t, response)
	assert.Contains(t, err.Error(), "context canceled")
}
