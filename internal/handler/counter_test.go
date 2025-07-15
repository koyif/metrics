package handler

import (
	"github.com/stretchr/testify/assert"
	"io"
	"net/http"
	"net/http/httptest"
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
				status: http.StatusBadRequest,
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

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := httptest.NewRequest(tt.when.method, tt.when.path, nil)
			if tt.when.metric != "" {
				r.SetPathValue("metric", tt.when.metric)
			}

			if tt.when.value != "" {
				r.SetPathValue("value", tt.when.value)
			}

			w := httptest.NewRecorder()

			handler.Handle(w, r)

			res := w.Result()
			defer res.Body.Close()

			assert.Equal(t, tt.want.status, res.StatusCode)

			if tt.want.body != "" {
				body, err := io.ReadAll(res.Body)

				assert.NoError(t, err)
				assert.JSONEq(t, tt.want.body, string(body))
			}
		})
	}
}
