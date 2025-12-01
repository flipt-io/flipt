package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/google/uuid"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

/* ---------- Domain Models ---------- */

type ApprovalRequest struct {
	ID            uuid.UUID       `json:"id"`
	SourceEnv     string          `json:"source_env"`
	TargetEnv     string          `json:"target_env"`
	SourceBranch  string          `json:"source_branch"`
	TargetBranch  string          `json:"target_branch"`
	RepoURL       string          `json:"repo_url"`
	SourceCommit  string          `json:"source_commit"`
	TargetCommit  string          `json:"target_commit"`
	ChangeType    string          `json:"change_type"`
	ChangePayload json.RawMessage `json:"change_payload"`
	Status        string          `json:"status"`
	ReviewState   string          `json:"review_state"`  // 👈 YENİ
	RequestedBy   string          `json:"requested_by"`
	CreatedAt     string          `json:"created_at"`
	UpdatedAt     string          `json:"updated_at"`
}


type LogEntry struct {
	Action    string `json:"action"`
	Actor     string `json:"actor"`
	Timestamp string `json:"timestamp"`
}

type User struct {
	Username string `json:"username"`
	FullName string `json:"full_name"`
	Role     string `json:"role"`
}

/* ---------- Roles ---------- */

const (
	RoleAdmin     = "ADMIN"
	RoleDeveloper = "DEVELOPER"
	RoleViewer    = "VIEWER"
)

/* ---------- Global Variables ---------- */

var (
	db         *sql.DB
	fliptURL   string
	fliptToken string
)

/* ---------- main ---------- */

func main() {
	// Load .env (optional in production)
	if err := godotenv.Load(); err != nil {
		log.Println("⚠️  .env not loaded (normal in production)")
	}

	// DB connection
	dbUser := os.Getenv("DB_USER")
	dbPass := os.Getenv("DB_PASS")
	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbName := os.Getenv("DB_NAME")

	connStr := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=disable",
		dbUser, dbPass, dbHost, dbPort, dbName,
	)

	var err error
	db, err = sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal("❌ DB connection error:", err)
	}

	if err := db.Ping(); err != nil {
		log.Fatal("❌ DB ping error:", err)
	}

	// Flipt config
	fliptURL = os.Getenv("FLIPT_URL")
	fliptToken = os.Getenv("FLIPT_TOKEN")
	if fliptURL == "" {
		log.Println("⚠️  FLIPT_URL not set, Flipt calls will not be made on approvals")
	} else {
		log.Println("✅ Flipt integration active. URL:", fliptURL)
	}

	// Seed default users (idempotent)
	if err := seedUsers(); err != nil {
		log.Fatal("❌ seedUsers error:", err)
	}

	fmt.Println("🚀 Backend running on http://localhost:8080")

	// Legacy routes (ilk versiyonla uyum için)
	http.HandleFunc("/approval-requests", withCORS(handleApprovalRequests))
	http.HandleFunc("/approval-requests/", withCORS(handleApprovalAction))
	http.HandleFunc("/approval-logs/", withCORS(handleApprovalLogs))
	http.HandleFunc("/me", withCORS(handleMe))

	// Yeni v1-style API routes
	http.HandleFunc("/api/v1/me", withCORS(handleMe))
	http.HandleFunc("/api/v1/changes", withCORS(handleApprovalRequests)) // GET/POST
	http.HandleFunc("/api/v1/changes/", withCORS(handleChangeRoutes))   // logs + approve/reject

	log.Fatal(http.ListenAndServe(":8080", nil))
}

/* ---------- Helpers / CORS / RBAC ---------- */

func withCORS(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, X-User")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}
		next(w, r)
	}
}

// Find current user from HTTP request
func getCurrentUser(r *http.Request) (*User, error) {
	username := r.Header.Get("X-User")
	if username == "" {
		return nil, errors.New("missing X-User header")
	}

	row := db.QueryRow(`
		SELECT username, full_name, role
		FROM users
		WHERE username = $1
	`, username)

	var u User
	if err := row.Scan(&u.Username, &u.FullName, &u.Role); err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("unknown user: %s", username)
		}
		return nil, err
	}
	return &u, nil
}

// Check user's role
func requireRole(u *User, allowedRoles ...string) error {
	for _, r := range allowedRoles {
		if u.Role == r {
			return nil
		}
	}
	return fmt.Errorf("forbidden: role %s is not allowed", u.Role)
}

// Example users at startup
func seedUsers() error {
	_, err := db.Exec(`
		INSERT INTO users (username, full_name, role) VALUES
		('yusuf',   'Yusuf Admin',     'ADMIN'),
		('dev1',    'Developer 1',     'DEVELOPER'),
		('viewer1', 'Read-only User',  'VIEWER')
		ON CONFLICT (username) DO NOTHING;
	`)
	return err
}

/* ---------- Flipt Integration Helper ---------- */

// Applies change_payload.new to Flipt
// change_payload format:
// {
//   "old": { ... },
//   "new": { "namespaceKey": "...", "key": "...", ... }
// }
func applyFliptChange(rawPayload json.RawMessage) error {
	if fliptURL == "" {
		// integration disabled: silently skip
		log.Println("ℹ️ FLIPT_URL empty, Flipt integration disabled (approve only updates DB)")
		return nil
	}

	var wrapper struct {
		Old json.RawMessage `json:"old"`
		New json.RawMessage `json:"new"`
	}
	if err := json.Unmarshal(rawPayload, &wrapper); err != nil {
		return fmt.Errorf("change_payload parse error: %w", err)
	}
	if len(wrapper.New) == 0 {
		return errors.New("change_payload.new is empty")
	}

	var flagMeta struct {
		NamespaceKey string `json:"namespaceKey"`
		Key          string `json:"key"`
	}
	if err := json.Unmarshal(wrapper.New, &flagMeta); err != nil {
		return fmt.Errorf("new flag meta parse error: %w", err)
	}
	if flagMeta.NamespaceKey == "" || flagMeta.Key == "" {
		return errors.New("new.namespaceKey or new.key missing")
	}

	url := fmt.Sprintf("%s/api/v1/namespaces/%s/flags/%s",
		fliptURL, flagMeta.NamespaceKey, flagMeta.Key)

	log.Println("ℹ️ Applying to Flipt:", url)

	req, err := http.NewRequest(http.MethodPut, url, bytes.NewReader(wrapper.New))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	if fliptToken != "" {
		req.Header.Set("Authorization", "Bearer "+fliptToken)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("flipt http error: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		return fmt.Errorf("flipt returned status %d", resp.StatusCode)
	}

	log.Println("✅ Flipt apply succeeded")
	return nil
}

/* ---------- Handlers ---------- */

// GET /me veya /api/v1/me
func handleMe(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	user, err := getCurrentUser(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}

// GET, POST /approval-requests
// GET, POST /api/v1/changes
func handleApprovalRequests(w http.ResponseWriter, r *http.Request) {
	switch r.Method {

	case http.MethodGet:
    rows, err := db.Query(`
        SELECT id,
               source_env,
               target_env,
               COALESCE(source_branch, ''),
               COALESCE(target_branch, ''),
               COALESCE(repo_url, ''),
               COALESCE(source_commit, ''),
               COALESCE(target_commit, ''),
               COALESCE(change_type, ''),
               change_payload,
               status,
               COALESCE(review_state, 'NONE'),
               requested_by,
               created_at,
               updated_at
        FROM approval_requests
        ORDER BY created_at DESC`)
    if err != nil {
        http.Error(w, err.Error(), 500)
        return
    }
    defer rows.Close()

    list := make([]ApprovalRequest, 0)
    for rows.Next() {
        var ar ApprovalRequest
        if err := rows.Scan(
            &ar.ID,
            &ar.SourceEnv,
            &ar.TargetEnv,
            &ar.SourceBranch,
            &ar.TargetBranch,
            &ar.RepoURL,
            &ar.SourceCommit,
            &ar.TargetCommit,
            &ar.ChangeType,
            &ar.ChangePayload,
            &ar.Status,
            &ar.ReviewState,
            &ar.RequestedBy,
            &ar.CreatedAt,
            &ar.UpdatedAt,
        ); err != nil {
            http.Error(w, err.Error(), 500)
            return
        }
        list = append(list, ar)
    }
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(list)


	case http.MethodPost:
		// Login required to create a request
		user, err := getCurrentUser(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}

		var ar ApprovalRequest
		if err := json.NewDecoder(r.Body).Decode(&ar); err != nil {
			http.Error(w, "Invalid JSON: "+err.Error(), 400)
			return
		}

		ar.ID = uuid.New()
		ar.Status = "PENDING"
		ar.RequestedBy = user.Username // body'den geleni değil, login kullanıcıyı kullan

		_, err = db.Exec(`
			INSERT INTO approval_requests
				(id,
				source_env,
				target_env,
				source_branch,
				target_branch,
				repo_url,
				source_commit,
				target_commit,
				change_type,
				change_payload,
				status,
				review_state,
				requested_by)
			VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13)
		`,
			ar.ID,
			ar.SourceEnv,
			ar.TargetEnv,
			ar.SourceBranch,
			ar.TargetBranch,
			ar.RepoURL,
			ar.SourceCommit,
			ar.TargetCommit,
			ar.ChangeType,
			ar.ChangePayload,
			ar.Status,
			"NONE",          // 👈 review_state başlangıçta NONE
			ar.RequestedBy,
		)

		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		// Audit log: CREATED
		db.Exec(`
			INSERT INTO approval_logs (request_id, action, actor)
			VALUES ($1, 'CREATED', $2)
		`, ar.ID, user.Username)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(ar)

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}


// POST /approval-requests/{id}/approve|reject
// POST /api/v1/changes/{id}/approve|reject
func handleApprovalAction(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	user, err := getCurrentUser(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}
	if err := requireRole(user, RoleAdmin); err != nil {
		http.Error(w, err.Error(), http.StatusForbidden)
		return
	}

	path := r.URL.Path

	// Prefix temizleme
	if strings.HasPrefix(path, "/approval-requests/") {
		path = strings.TrimPrefix(path, "/approval-requests/")
	} else if strings.HasPrefix(path, "/api/v1/changes/") {
		path = strings.TrimPrefix(path, "/api/v1/changes/")
	} else {
		http.Error(w, "Invalid path", http.StatusBadRequest)
		return
	}

	// Beklenen format: {uuid} veya {uuid}/approve|reject
	parts := strings.Split(path, "/")
	if len(parts) == 0 {
		http.Error(w, "Invalid path", http.StatusBadRequest)
		return
	}

	id := parts[0]
	action := ""
	if len(parts) > 1 {
		action = parts[1]
	}

	if action != "approve" && action != "reject" {
		http.Error(w, "Invalid action", http.StatusBadRequest)
		return
	}

	status := "APPROVED"
	if action == "reject" {
		status = "REJECTED"
	}

	// İlgili isteğin payload'ını çek
	var changePayload json.RawMessage
	var reviewState string
	err = db.QueryRow(`
		SELECT change_payload, COALESCE(review_state, 'NONE')
		FROM approval_requests
		WHERE id = $1
	`, id).Scan(&changePayload, &reviewState)
	if err != nil {
		http.Error(w, "request not found: "+err.Error(), 404)
		return
	}

// APPROVE ise ve review yapılmamışsa hata ver
if action == "approve" && reviewState != "REVIEWED" {
    http.Error(w, "cannot approve before review", http.StatusBadRequest)
    return
}


	// APPROVE ise Flipt'e apply etmeyi dene (best-effort)
	if action == "approve" {
		if err := applyFliptChange(changePayload); err != nil {
			log.Println("❌ Flipt apply error (approve continues):", err)
		}
	}

	_, err = db.Exec(`
		UPDATE approval_requests
		SET status=$1, updated_at=NOW()
		WHERE id=$2
	`, status, id)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	// Audit log: APPROVED / REJECTED
	db.Exec(`
		INSERT INTO approval_logs (request_id, action, actor)
		VALUES ($1, $2, $3)
	`, id, status, user.Username)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": status})
}

// POST /api/v1/changes/{id}/review
func handleReviewAction(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPost {
        http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
        return
    }

    user, err := getCurrentUser(r)
    if err != nil {
        http.Error(w, err.Error(), http.StatusUnauthorized)
        return
    }

    // Reviewer rolü: DEVELOPER veya ADMIN
    if err := requireRole(user, RoleAdmin, RoleDeveloper); err != nil {
        http.Error(w, err.Error(), http.StatusForbidden)
        return
    }

    path := r.URL.Path
    // /api/v1/changes/{id}/review
    trimmed := strings.TrimPrefix(path, "/api/v1/changes/")
    trimmed = strings.TrimSuffix(trimmed, "/review")
    id := strings.Trim(trimmed, "/")
    if id == "" {
        http.Error(w, "Missing ID", http.StatusBadRequest)
        return
    }

    // review_state'i güncelle
    _, err = db.Exec(`
        UPDATE approval_requests
        SET review_state = 'REVIEWED', updated_at = NOW()
        WHERE id = $1
    `, id)
    if err != nil {
        http.Error(w, err.Error(), 500)
        return
    }

    // Audit log
    db.Exec(`
        INSERT INTO approval_logs (request_id, action, actor)
        VALUES ($1, $2, $3)
    `, id, "REVIEWED", user.Username)

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]string{
        "review_state": "REVIEWED",
    })
}


// /api/v1/changes/{id}/... router'ı
func handleChangeRoutes(w http.ResponseWriter, r *http.Request) {
    // /api/v1/changes/{id}/review
    if strings.HasPrefix(r.URL.Path, "/api/v1/changes/") && strings.HasSuffix(r.URL.Path, "/review") {
        handleReviewAction(w, r)
        return
    }

    // /api/v1/changes/{id}/logs
    if strings.HasPrefix(r.URL.Path, "/api/v1/changes/") && strings.HasSuffix(r.URL.Path, "/logs") {
        handleApprovalLogs(w, r)
        return
    }

    // Geri kalan her şey approve/reject olarak değerlendirilir
    handleApprovalAction(w, r)
}


// GET /approval-logs/{id}
// GET /api/v1/changes/{id}/logs
func handleApprovalLogs(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	path := r.URL.Path
	var id string

	if strings.HasPrefix(path, "/approval-logs/") {
		id = strings.TrimPrefix(path, "/approval-logs/")
	} else if strings.HasPrefix(path, "/api/v1/changes/") && strings.HasSuffix(path, "/logs") {
		// /api/v1/changes/{id}/logs
		trimmed := strings.TrimPrefix(path, "/api/v1/changes/")
		trimmed = strings.TrimSuffix(trimmed, "/logs")
		id = strings.Trim(trimmed, "/")
	} else {
		http.Error(w, "Invalid path", http.StatusBadRequest)
		return
	}

	if id == "" {
		http.Error(w, "Missing ID", 400)
		return
	}

	rows, err := db.Query(`
		SELECT action, actor, timestamp
		FROM approval_logs
		WHERE request_id=$1
		ORDER BY timestamp ASC
	`, id)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	defer rows.Close()

	var logs []LogEntry
	for rows.Next() {
		var l LogEntry
		if err := rows.Scan(&l.Action, &l.Actor, &l.Timestamp); err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		logs = append(logs, l)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(logs)
}
