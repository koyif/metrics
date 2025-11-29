package health_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"

	"github.com/koyif/metrics/internal/handler/health"
	"github.com/koyif/metrics/internal/repository"
	"github.com/koyif/metrics/internal/service"
)

// Example_healthCheck demonstrates how to check the service health status.
// The health endpoint verifies that the service is running and the database
// (if configured) is reachable.
func Example_healthCheck() {
	// Create test dependencies
	repo := repository.NewMetricsRepository()
	svc := service.NewMetricsService(repo, nil)
	handler := health.NewPingHandler(svc)

	// Create test server
	ts := httptest.NewServer(http.HandlerFunc(handler.Handle))
	defer ts.Close()

	// Send GET request
	resp, err := http.Get(ts.URL)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	defer resp.Body.Close()

	fmt.Printf("Status: %d\n", resp.StatusCode)
	// Output: Status: 200
}
