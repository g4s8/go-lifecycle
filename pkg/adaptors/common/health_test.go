package common_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/g4s8/go-lifecycle/pkg/adaptors/common"
)

func waitReady(t *testing.T, url string) {
	t.Helper()
	for i := 0; i < 50; i++ {
		resp, err := http.Get(url)
		if err == nil {
			resp.Body.Close()
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatalf("server at %s did not become ready", url)
}

func startHealthRunner(t *testing.T, port int, path string, checker func(ctx context.Context) []error) {
	t.Helper()
	runner := common.NewHealthRunner(port, path, checker)
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)
	go runner.Run(ctx) //nolint:errcheck
}

func TestHealthRunner_Healthy(t *testing.T) {
	port := freePort(t)
	startHealthRunner(t, port, "/healthz", func(_ context.Context) []error {
		return nil
	})
	url := fmt.Sprintf("http://127.0.0.1:%d/healthz", port)
	waitReady(t, url)

	resp, err := http.Get(url)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	var data map[string]interface{}
	require.NoError(t, json.Unmarshal(body, &data))
	assert.Equal(t, "healthy", data["status"])
	assert.Nil(t, data["errors"])
}

func TestHealthRunner_Unhealthy(t *testing.T) {
	port := freePort(t)
	startHealthRunner(t, port, "/healthz", func(_ context.Context) []error {
		return []error{errors.New("db down"), errors.New("cache unreachable")}
	})
	url := fmt.Sprintf("http://127.0.0.1:%d/healthz", port)
	waitReady(t, url)

	resp, err := http.Get(url)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusServiceUnavailable, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	var data map[string]interface{}
	require.NoError(t, json.Unmarshal(body, &data))
	assert.Equal(t, "unhealthy", data["status"])
	errs, ok := data["errors"].([]interface{})
	require.True(t, ok)
	assert.Len(t, errs, 2)
}

func TestHealthRunner_CustomPath(t *testing.T) {
	port := freePort(t)
	startHealthRunner(t, port, "/ready", func(_ context.Context) []error {
		return nil
	})
	url := fmt.Sprintf("http://127.0.0.1:%d/ready", port)
	waitReady(t, url)

	resp, err := http.Get(url)
	require.NoError(t, err)
	resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	resp2, err := http.Get(fmt.Sprintf("http://127.0.0.1:%d/healthz", port))
	require.NoError(t, err)
	resp2.Body.Close()
	assert.Equal(t, http.StatusNotFound, resp2.StatusCode)
}
