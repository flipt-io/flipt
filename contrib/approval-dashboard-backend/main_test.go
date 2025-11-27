package main

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestHandleApprovalRequests_Get(t *testing.T) {
	var mock sqlmock.Sqlmock
	var err error
	db, mock, err = sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New error: %v", err)
	}
	defer db.Close()

	rows := sqlmock.NewRows([]string{
		"id", "source_env", "target_env", "change_payload",
		"status", "requested_by", "created_at", "updated_at",
	}).AddRow(
		"11111111-1111-1111-1111-111111111111",
		"dev",
		"staging",
		[]byte(`{"old":{"enabled":false},"new":{"enabled":true}}`),
		"PENDING",
		"yusuf",
		"2025-01-01T00:00:00Z",
		"2025-01-01T00:00:00Z",
	)

	mock.ExpectQuery("SELECT id, source_env, target_env").
		WillReturnRows(rows)

	req := httptest.NewRequest(http.MethodGet, "/approval-requests", nil)
	w := httptest.NewRecorder()

	handleApprovalRequests(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("SQL expectations not met: %v", err)
	}
}

func TestHandleApprovalRequests_Post(t *testing.T) {
	var mock sqlmock.Sqlmock
	var err error
	db, mock, err = sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New error: %v", err)
	}
	defer db.Close()

	mock.ExpectExec("INSERT INTO approval_requests").
		WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectExec("INSERT INTO approval_logs").
		WillReturnResult(sqlmock.NewResult(1, 1))

	body := []byte(`{
		"source_env": "dev",
		"target_env": "staging",
		"change_payload": {
			"old": { "enabled": false },
			"new": { "enabled": true }
		},
		"requested_by": "yusuf"
	}`)

	req := httptest.NewRequest(http.MethodPost, "/approval-requests", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	handleApprovalRequests(w, req)

	if w.Code != http.StatusCreated && w.Code != http.StatusOK {
		t.Fatalf("expected 201/200, got %d", w.Code)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("SQL expectations not met: %v", err)
	}
}
