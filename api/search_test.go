package api

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/davecgh/go-spew/spew"
	"github.com/enigmaquip/gonab/db"
)

func TestSearch(t *testing.T) {
	dbh := db.NewMemoryDBHandle(false, false)
	n := configRoutes(dbh)

	req, err := http.NewRequest("GET", "/gonab/api?t=search&q=foo&apikey=123", nil)
	if err != nil {
		t.Fatalf("Error setting up request: %s", err)
	}
	respRec := httptest.NewRecorder()
	n.ServeHTTP(respRec, req)

	if respRec.Code != http.StatusOK {
		spew.Dump(respRec)
		t.Fatalf("Error running caps api: %d", respRec.Code)
	}
}
