package main

import (
	"bytes"
	"testing"
	"net/http"
	"net/http/httptest"

	"github.com/gin-gonic/gin"
	"go.b8s.dev/rollingpin/config"
)

func TestAuthSuccess(t *testing.T) {
	conf := &config.Config{
		AuthToken: "test1234",
	}
	req, _ := http.NewRequest("POST", "/", bytes.NewBufferString(""))
	req.Header.Add("authorization", "Bearer test1234")
	ctx, resp := buildTestConn(req)
	auth(conf)(ctx)
	if resp.Code != 200 {
		t.Errorf("Request should have been 200 but was %d!", resp.Code)
	}
}

func TestAuthMismatch(t *testing.T) {
	conf := &config.Config{
		AuthToken: "asdfljasfdadsfjasdf",
	}
	req, _ := http.NewRequest("POST", "/", bytes.NewBufferString(""))
	req.Header.Add("authorization", "Bearer test1234")
	ctx, resp := buildTestConn(req)
	auth(conf)(ctx)
	if resp.Code != 401 {
		t.Errorf("Request should have been 401 but was %d!", resp.Code)
	}
}

func buildTestConn(req *http.Request) (*gin.Context, *httptest.ResponseRecorder) {
        w := httptest.NewRecorder()
        conn, _ := gin.CreateTestContext(w)
        conn.Request = req
        return conn, w
}
