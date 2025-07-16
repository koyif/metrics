package handler

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
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

	handler := NewCountersHandler(MockCountersRepository{})
	mux := http.NewServeMux()
	mux.HandleFunc("/update/counter/{metric}/{value}", handler.Handle)
	mux.HandleFunc("/update/counter/{metric}", handler.Handle)
	mux.HandleFunc("/update/counter", handler.Handle)

	server := httptest.NewServer(mux)
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
			require.NoError(t, err)

			defer response.Body.Close()

			assert.Equal(t, tt.want.status, response.StatusCode)

			if tt.want.body != "" {
				body, err := io.ReadAll(response.Body)

				assert.NoError(t, err)
				assert.JSONEq(t, tt.want.body, string(body))
			}
		})
	}
}
