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

	"github.com/google/uuid"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

/* ---------- Domain Modelleri ---------- */

type ApprovalRequest struct {
	ID            uuid.UUID       `json:"id"`
	SourceEnv     string          `json:"source_env"`
	TargetEnv     string          `json:"target_env"`
	ChangePayload json.RawMessage `json:"change_payload"`
	Status        string          `json:"status"`
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

/* ---------- Roller ---------- */

const (
	RoleAdmin     = "ADMIN"
	RoleDeveloper = "DEVELOPER"
	RoleViewer    = "VIEWER"
)

/* ---------- Global Değişkenler ---------- */

var (
	db         *sql.DB
	fliptURL   string
	fliptToken string
)

/* ---------- main ---------- */

func main() {
	// .env yükle
	if err := godotenv.Load(); err != nil {
		log.Println("⚠️  .env yüklenemedi (production'da normaldir)")
	}

	// DB bağlantısı
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
		log.Fatal("❌ DB bağlantı hatası:", err)
	}

	// Flipt config
	fliptURL = os.Getenv("FLIPT_URL")
	fliptToken = os.Getenv("FLIPT_TOKEN")
	if fliptURL == "" {
		log.Println("⚠️  FLIPT_URL tanımlı değil, approval onaylarında Flipt çağrısı yapılmayacak")
	} else {
		log.Println("✅ Flipt entegrasyonu aktif. URL:", fliptURL)
	}

	// Varsayılan kullanıcıları seed et (idempotent)
	if err := seedUsers(); err != nil {
		log.Fatal("❌ seedUsers hatası:", err)
	}

	fmt.Println("🚀 Backend running on http://localhost:8080")

	http.HandleFunc("/approval-requests", withCORS(handleApprovalRequests))
	http.HandleFunc("/approval-requests/", withCORS(handleApprovalAction))
	http.HandleFunc("/approval-logs/", withCORS(handleApprovalLogs))
	http.HandleFunc("/me", withCORS(handleMe))

	log.Fatal(http.ListenAndServe(":8080", nil))
}

/* ---------- Yardımcılar / CORS / RBAC ---------- */

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

// HTTP isteğinden aktif kullanıcıyı bul
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

// Kullanıcının rolünü kontrol et
func requireRole(u *User, allowedRoles ...string) error {
	for _, r := range allowedRoles {
		if u.Role == r {
			return nil
		}
	}
	return fmt.Errorf("forbidden: role %s is not allowed", u.Role)
}

// Başlangıçta örnek kullanıcılar
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

/* ---------- Flipt Entegrasyon Helper ---------- */

// change_payload.new içeriğini Flipt'e uygular
// change_payload formatı:
// {
//   "old": { ... },
//   "new": { "namespaceKey": "...", "key": "...", ... }
// }
func applyFliptChange(rawPayload json.RawMessage) error {
    if fliptURL == "" {
        // entegrasyon kapalı: sessizce geç
        log.Println("ℹ️ FLIPT_URL boş, Flipt entegrasyonu devre dışı (approve sadece DB'de güncellendi)")
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
        return errors.New("change_payload.new boş")
    }

    var flagMeta struct {
        NamespaceKey string `json:"namespaceKey"`
        Key          string `json:"key"`
    }
    if err := json.Unmarshal(wrapper.New, &flagMeta); err != nil {
        return fmt.Errorf("new flag meta parse error: %w", err)
    }
    if flagMeta.NamespaceKey == "" || flagMeta.Key == "" {
        return errors.New("new.namespaceKey veya new.key eksik")
    }

    url := fmt.Sprintf("%s/api/v1/namespaces/%s/flags/%s",
        fliptURL, flagMeta.NamespaceKey, flagMeta.Key)

    log.Println("ℹ️ Flipt'e apply ediliyor:", url)

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

    log.Println("✅ Flipt apply başarılı")
    return nil
}

/* ---------- Handlers ---------- */

// GET /me -> aktif kullanıcı bilgisi
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

	json.NewEncoder(w).Encode(user)
}

// GET, POST /approval-requests
func handleApprovalRequests(w http.ResponseWriter, r *http.Request) {
	switch r.Method {

	case http.MethodGet:
		// Listeleme: herkes görebilir (istersen burada da rol kontrolü ekleyebiliriz)
		rows, err := db.Query(`
			SELECT id, source_env, target_env, change_payload, status, requested_by, created_at, updated_at
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
				&ar.ID, &ar.SourceEnv, &ar.TargetEnv, &ar.ChangePayload,
				&ar.Status, &ar.RequestedBy, &ar.CreatedAt, &ar.UpdatedAt,
			); err != nil {
				http.Error(w, err.Error(), 500)
				return
			}
			list = append(list, ar)
		}
		json.NewEncoder(w).Encode(list)

	case http.MethodPost:
		// Request oluşturmak için login zorunlu
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
		ar.RequestedBy = user.Username // Body'den gelene güvenmiyoruz

		_, err = db.Exec(`
			INSERT INTO approval_requests (id, source_env, target_env, change_payload, status, requested_by)
			VALUES ($1,$2,$3,$4,$5,$6)
		`, ar.ID, ar.SourceEnv, ar.TargetEnv, ar.ChangePayload, ar.Status, ar.RequestedBy)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		// Audit log: CREATED
		db.Exec(`
			INSERT INTO approval_logs (request_id, action, actor)
			VALUES ($1, 'CREATED', $2)
		`, ar.ID, user.Username)

		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(ar)

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// POST /approval-requests/{id}/approve veya reject
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

	// /approval-requests/{id}/{action}
	path := r.URL.Path[len("/approval-requests/"):]
	if len(path) < 36 {
		http.Error(w, "Invalid ID", 400)
		return
	}

	id := path[:36]
	action := ""
	if len(path) > 36 {
		// "uuid/approve" -> 36 + 1 + action
		action = path[37:]
	}

	if action != "approve" && action != "reject" {
		http.Error(w, "Invalid action", 400)
		return
	}

	status := "APPROVED"
	if action == "reject" {
		status = "REJECTED"
	}

	// İlgili request'in payload'unu al
	var changePayload json.RawMessage
	err = db.QueryRow(`
		SELECT change_payload
		FROM approval_requests
		WHERE id = $1
	`, id).Scan(&changePayload)
	if err != nil {
		http.Error(w, "request not found: "+err.Error(), 404)
		return
	}

	// APPROVE ise Flipt'e uygula (best-effort)
	if action == "approve" {
		if err := applyFliptChange(changePayload); err != nil {
			// Sadece log atalım, kullanıcıya 500 dönmeyelim
			log.Println("❌ Flipt apply error (approve yine de devam ediyor):", err)
			// İstersen ileride buraya özel bir field ekleyip UI'da "Flipt'e apply edilemedi" diye gösterebiliriz.
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

	json.NewEncoder(w).Encode(map[string]string{"status": status})
}

// GET /approval-logs/{id}
func handleApprovalLogs(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	id := r.URL.Path[len("/approval-logs/"):]
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

	json.NewEncoder(w).Encode(logs)
}
