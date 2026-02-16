-- =========================================================
-- OpsLens Control Plane - PostgreSQL
-- =========================================================

CREATE EXTENSION IF NOT EXISTS "pgcrypto";
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- =========================================================
-- TENANTS
-- =========================================================
CREATE TABLE tenants (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT NOT NULL,
    slug TEXT UNIQUE NOT NULL,
    plan TEXT DEFAULT 'free',
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT now(),
    updated_at TIMESTAMP DEFAULT now()
);

-- =========================================================
-- USERS
-- =========================================================
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    username TEXT UNIQUE NOT NULL,
    email TEXT UNIQUE NOT NULL,
    password_hash TEXT NOT NULL,
    is_active BOOLEAN DEFAULT true,
    is_super_admin BOOLEAN DEFAULT false,
    created_at TIMESTAMP DEFAULT now(),
    updated_at TIMESTAMP DEFAULT now()
);

-- =========================================================
-- USER ↔ TENANT MAPPING (Multi-Tenant SaaS Model)
-- =========================================================
CREATE TABLE user_tenants (
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    tenant_id UUID REFERENCES tenants(id) ON DELETE CASCADE,
    joined_at TIMESTAMP DEFAULT now(),
    PRIMARY KEY (user_id, tenant_id)
);

CREATE INDEX idx_user_tenants_user ON user_tenants(user_id);
CREATE INDEX idx_user_tenants_tenant ON user_tenants(tenant_id);

-- =========================================================
-- ROLES & PERMISSIONS
-- =========================================================
CREATE TABLE roles (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID REFERENCES tenants(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    description TEXT,
    UNIQUE (tenant_id, name)
);

CREATE TABLE permissions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT UNIQUE NOT NULL,
    description TEXT
);

CREATE TABLE role_permissions (
    role_id UUID REFERENCES roles(id) ON DELETE CASCADE,
    permission_id UUID REFERENCES permissions(id) ON DELETE CASCADE,
    PRIMARY KEY (role_id, permission_id)
);

CREATE TABLE groups (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID REFERENCES tenants(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    UNIQUE (tenant_id, name)
);

CREATE TABLE group_roles (
    group_id UUID REFERENCES groups(id) ON DELETE CASCADE,
    role_id UUID REFERENCES roles(id) ON DELETE CASCADE,
    PRIMARY KEY (group_id, role_id)
);

CREATE TABLE user_groups (
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    group_id UUID REFERENCES groups(id) ON DELETE CASCADE,
    PRIMARY KEY (user_id, group_id)
);

-- =========================================================
-- API KEYS
-- =========================================================
CREATE TABLE api_keys (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID REFERENCES tenants(id) ON DELETE CASCADE,
    name TEXT,
    key_hash TEXT NOT NULL,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT now(),
    last_used_at TIMESTAMP
);

CREATE INDEX idx_api_keys_tenant ON api_keys(tenant_id);

-- =========================================================
-- SESSIONS
-- =========================================================
CREATE TABLE sessions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    refresh_token_hash TEXT NOT NULL,
    expires_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP DEFAULT now()
);

CREATE INDEX idx_sessions_user ON sessions(user_id);

-- =========================================================
-- AUDIT LOGS
-- =========================================================
CREATE TABLE audit_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID REFERENCES tenants(id),
    user_id UUID REFERENCES users(id),
    action TEXT NOT NULL,
    resource TEXT,
    metadata JSONB,
    created_at TIMESTAMP DEFAULT now()
);

CREATE INDEX idx_audit_tenant ON audit_logs(tenant_id);
CREATE INDEX idx_audit_created_at ON audit_logs(created_at);

-- =========================================================
-- BILLING
-- =========================================================

CREATE TABLE IF NOT EXISTS billing_plans (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT,
    metrics_price_per_million NUMERIC,
    logs_price_per_gb NUMERIC,
    retention_days INTEGER
);

CREATE TABLE IF NOT EXISTS invoices (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID REFERENCES tenants(id),
    period_start DATE,
    period_end DATE,
    total_amount NUMERIC,
    status TEXT DEFAULT 'pending',
    created_at TIMESTAMP DEFAULT now()
);

CREATE TABLE IF NOT EXISTS invoice_items (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    invoice_id UUID REFERENCES invoices(id) ON DELETE CASCADE,
    description TEXT,
    quantity NUMERIC,
    unit_price NUMERIC,
    amount NUMERIC
);

-- =========================================================
-- ALERT RULE ENGINE
-- =========================================================

CREATE TABLE IF NOT EXISTS alert_rules (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID REFERENCES tenants(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    metric_name TEXT,
    condition TEXT,
    threshold NUMERIC,
    duration_seconds INTEGER,
    severity TEXT,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT now()
);

CREATE TABLE IF NOT EXISTS alert_instances (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    rule_id UUID REFERENCES alert_rules(id) ON DELETE CASCADE,
    tenant_id UUID,
    agent_id TEXT,
    triggered_at TIMESTAMP DEFAULT now(),
    resolved_at TIMESTAMP,
    status TEXT DEFAULT 'firing'
);

CREATE INDEX IF NOT EXISTS idx_alert_rules_tenant
ON alert_rules(tenant_id);


-- =========================================================
-- DEFAULT PERMISSIONS
-- =========================================================
INSERT INTO permissions (name, description) VALUES
('hosts.read', 'View hosts'),
('hosts.write', 'Modify hosts'),
('logs.read', 'View logs'),
('logs.delete', 'Delete logs'),
('users.read', 'View users'),
('users.write', 'Manage users'),
('api_keys.manage', 'Manage API keys'),
('alerts.manage', 'Manage alerts')
ON CONFLICT DO NOTHING;

-- =========================================================
-- DEFAULT TENANT
-- =========================================================
INSERT INTO tenants (name, slug, plan)
VALUES ('Default Tenant', 'default-tenant', 'enterprise')
ON CONFLICT DO NOTHING;

-- =========================================================
-- DEFAULT ADMIN ROLE
-- =========================================================
INSERT INTO roles (tenant_id, name, description)
SELECT id, 'admin', 'Full access role'
FROM tenants
WHERE slug = 'default-tenant'
ON CONFLICT DO NOTHING;

INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r
CROSS JOIN permissions p
JOIN tenants t ON r.tenant_id = t.id
WHERE r.name = 'admin'
AND t.slug = 'default-tenant'
ON CONFLICT DO NOTHING;

-- =========================================================
-- DEFAULT ADMIN GROUP
-- =========================================================
INSERT INTO groups (tenant_id, name)
SELECT id, 'administrators'
FROM tenants
WHERE slug = 'default-tenant'
ON CONFLICT DO NOTHING;

INSERT INTO group_roles (group_id, role_id)
SELECT g.id, r.id
FROM groups g
JOIN roles r ON g.tenant_id = r.tenant_id
JOIN tenants t ON g.tenant_id = t.id
WHERE g.name = 'administrators'
AND r.name = 'admin'
AND t.slug = 'default-tenant'
ON CONFLICT DO NOTHING;

-- =========================================================
-- DEFAULT ADMIN USER (admin/admin)
-- =========================================================
INSERT INTO users (username, email, password_hash, is_super_admin)
VALUES (
    'admin',
    'admin@opslens.local',
    '$2a$12$SbDpphIeXQFOffc9DieDh.iF0YY25IiFfAlbTb2w8/rVkrGPSDv8m',
    true
)
ON CONFLICT (username) DO NOTHING;

INSERT INTO user_tenants (user_id, tenant_id)
SELECT u.id, t.id
FROM users u
JOIN tenants t ON t.slug = 'default-tenant'
WHERE u.username = 'admin'
ON CONFLICT DO NOTHING;

INSERT INTO user_groups (user_id, group_id)
SELECT u.id, g.id
FROM users u
JOIN groups g ON g.name = 'administrators'
JOIN tenants t ON g.tenant_id = t.id
WHERE u.username = 'admin'
AND t.slug = 'default-tenant'
ON CONFLICT DO NOTHING;
