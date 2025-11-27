-- 002_rbac.sql
-- Basit ama gerçekçi bir RBAC modeli

CREATE TABLE IF NOT EXISTS users (
    username   TEXT PRIMARY KEY,
    full_name  TEXT,
    role       TEXT NOT NULL CHECK (role IN ('ADMIN', 'DEVELOPER', 'VIEWER')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- approval_requests.requested_by -> users.username FK
ALTER TABLE approval_requests
    ADD CONSTRAINT IF NOT EXISTS fk_approval_requests_user
    FOREIGN KEY (requested_by)
    REFERENCES users(username)
    ON UPDATE CASCADE
    ON DELETE RESTRICT;

-- Örnek kullanıcılar
INSERT INTO users (username, full_name, role) VALUES
('yusuf',   'Yusuf Admin',     'ADMIN'),
('dev1',    'Developer 1',     'DEVELOPER'),
('viewer1', 'Read-only User',  'VIEWER')
ON CONFLICT (username) DO NOTHING;
