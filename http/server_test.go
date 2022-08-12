package http

import (
	"bytes"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/echovl/orderflo-dev/layerhub"
	"go.uber.org/zap"
)

func setupTestServer(t *testing.T) *Server {
	logger, _ := zap.NewProduction()
	defer logger.Sync()
	return NewServer(Config{
		Core: layerhub.New(layerhub.CoreConfig{
			Logger:          logger,
			Uploader:        nil,
			Pixabay:         nil,
			Pexels:          nil,
			PaymentProvider: nil,
			Renderer:        nil,
		}),
	})
}

func TestRequestParser(t *testing.T) {
	sv := setupTestServer(t)

	type request struct {
		BodyField  int    `json:"bf" validate:"max=20"`
		QueryField int    `query:"qf" json:"qf"`
		ErrorMsg   string `json:"msg"`
	}

	handler := func(c *fiber.Ctx) error {
		var req request
		err := sv.requestParser(c, &req)
		if err != nil {
			return c.JSON(request{ErrorMsg: err.Error()})
		}
		return c.JSON(req)
	}

	sv.App.Post("/test", handler)
	sv.App.Get("/test", handler)

	testcases := []struct {
		name    string
		body    string
		method  string
		url     string
		want    request
		wantErr bool
	}{
		{"body", `{"bf":10}`, http.MethodPost, "http://localhost/test", request{10, 0, ""}, false},
		{"queryparams", "", http.MethodGet, "http://localhost/test?qf=10", request{0, 10, ""}, false},
		{"body and queryparams", `{"bf":10}`, http.MethodPost, "http://localhost/test?qf=10", request{10, 10, ""}, false},
		{"failed validation", `{"bf":100}`, http.MethodPost, "http://localhost/test", request{}, true},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			body := bytes.NewBufferString(tc.body)
			req, err := http.NewRequest(tc.method, tc.url, body)
			if err != nil {
				t.Fatal(err)
			}

			if tc.method == http.MethodPost {
				req.Header.Add("Content-Type", "application/json")
			}

			resp, err := sv.App.Test(req)
			if err != nil {
				t.Fatal(err)
			}
			defer resp.Body.Close()

			var got request
			err = json.NewDecoder(resp.Body).Decode(&got)
			if err != nil {
				t.Errorf("decoding request: %s", err)
			}

			if tc.wantErr {
				if got.ErrorMsg == "" {
					t.Errorf("validation should fail: got: %v", got)
				}
			} else {
				if got != tc.want {
					t.Errorf("mismatched requests:\ngot: %v\nwant: %v", got, tc.want)
				}
			}
		})
	}
}
