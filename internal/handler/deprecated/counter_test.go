package deprecated

import (
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type MockCountersRepository struct{}

func (MockCountersRepository) StoreCounter(metricName string, value int64) error {
	return nil
}

func TestCountersHandler_Handle(t *testing.T) {
	type when struct {
		method string
		path   string
		metric string
		value  string
	}
	type want struct {
		status int
		body   string
	}
	tests := []struct {
		name string
		when when
		want want
	}{
		{
			name: "OK",
			when: when{
				method: http.MethodPost,
				path:   "/update/counter",
				metric: "counter",
				value:  "1",
			},
			want: want{
				status: http.StatusOK,
			},
		},
		{
			name: "no metric name",
			when: when{
				method: http.MethodPost,
				path:   "/update/counter",
				value:  "1",
			},
			want: want{
				status: http.StatusNotFound,
			},
		},
		{
			name: "invalid http method PUT",
			when: when{
				method: http.MethodPut,
				path:   "/update/counter",
				metric: "counter",
				value:  "1",
			},
			want: want{
				status: http.StatusMethodNotAllowed,
			},
		},
		{
			name: "invalid http method GET",
			when: when{
				method: http.MethodGet,
				path:   "/update/counter",
				metric: "counter",
				value:  "1",
			},
			want: want{
				status: http.StatusMethodNotAllowed,
			},
		},
		{
			name: "no value",
			when: when{
				method: http.MethodPost,
				path:   "/update/counter",
				metric: "counter",
			},
			want: want{
				status: http.StatusNotFound,
			},
		},
		{
			name: "invalid value float",
			when: when{
				method: http.MethodPost,
				path:   "/update/counter",
				metric: "counter",
				value:  "1.1",
			},
			want: want{
				status: http.StatusBadRequest,
			},
		},
		{
			name: "invalid value string",
			when: when{
				method: http.MethodPost,
				path:   "/update/counter",
				metric: "counter",
				value:  "string",
			},
			want: want{
				status: http.StatusBadRequest,
			},
		},
	}

	handler := NewCountersPostHandler(MockCountersRepository{})

	r := chi.NewRouter()
	r.Post("/update/counter/{metric}/{value}", handler.Handle)

	server := httptest.NewServer(r)
	defer server.Close()

	baseURL, err := url.Parse(server.URL)
	assert.NoError(t, err)

	client := http.Client{}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reqURL := baseURL.
				JoinPath(tt.when.path).
				JoinPath(tt.when.metric).
				JoinPath(tt.when.value).
				String()
			req, err := http.NewRequest(tt.when.method, reqURL, nil)
			require.NoError(t, err)

			response, err := client.Do(req)
			if err != nil {
				err := response.Body.Close()
				if err != nil {
					slog.Error(fmt.Sprintf("Failed to close response body: %s", err))
					return
				}
			}

			require.NoError(t, err)

			defer func(Body io.ReadCloser) {
				err := Body.Close()
				if err != nil {
					t.Errorf("error closing response body: %v", err)
				}
			}(response.Body)

			assert.Equal(t, tt.want.status, response.StatusCode)

			if tt.want.body != "" {
				body, err := io.ReadAll(response.Body)

				require.NoError(t, err)
				assert.JSONEq(t, tt.want.body, string(body))
			}
		})
	}
}
